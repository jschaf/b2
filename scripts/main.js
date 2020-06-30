// Detect adblock.
//
// Find out how common adblock is. Sets window.adblockStatus to a string of:
// 'unknown', 'active', or 'inactive'.
(async () => {
  const intervalMs = 51;
  const testRuns = 4;

  window.adblockStatus = 'unknown';
  for (let i = 0; i < testRuns; i++) {
    const elem = window.document.getElementById('banner_ad');
    const isBlocked = elem == null ||
        elem.offsetHeight === 0 ||
        elem.style.display === 'none' ||
        elem.style.visibility === 'hidden';
    if (isBlocked) {
      window.adblockStatus = 'blocking';
      console.debug(`adblock status: ${window.adblockStatus}`)
      return;
    }
    await new Promise(resolve => setTimeout(resolve, intervalMs));
  }
  window.adblockStatus = 'none';
  console.debug(`adblock status: ${window.adblockStatus}`)
})();

// Load heap stubs that run while the real heap.js downloads.
(() => {
  window.heap = [];
  const stubHeapFn = (fnName) => (...args) => {
    window.heap.push([fnName, ...args]);
  };
  const fnNames = [
    'addEventProperties',
    'addUserProperties',
    'clearEventProperties',
    'identify',
    'resetIdentity',
    'removeEventProperty',
    'setEventProperties',
    'track',
    'unsetEventProperty'
  ];
  for (const s of fnNames) {
    window.heap[s] = stubHeapFn(s)
  }
})();

/**
 * PreviewLifecycle manages the state transitions for the preview box display.
 * The transitions are complex, resulting from interactions between the target
 * link and preview div. Use-cases we want to support:
 *
 * - Continue showing the preview when moving from the target to preview box
 *   for a grace period.
 *
 * - Continue showing the preview when leaving the box but quickly returning for
 *   a grace period.
 *
 * - Choosing whether to display the preview box above or below the target. We
 *   generally prefer above to avoid blocking lines the user will read.
 *
 * - Dynamically generating preview box content for things like citation hovers.
 *   We dynamically generate previews when the content exists on the current
 *   page.
 */
class PreviewLifecycle {
  constructor() {
    /**
     * The current, displayed preview target pending or displayed. If no preview
     * is displayed, currentTarget is null.
     * @type {?HTMLElement}
     */
    this.currentTarget = null;
    /** @type {number} */
    this.showPreviewTimer = 0;
    /** @type {number} */
    this.hidePreviewTimer = 0;
    /**
     * A singleton div element to hold previews of preview target links.
     * Lazily initialized on the first hover of a preview target. Contains
     * contentEl.
     * @type {?HTMLElement}
     */
    this.boxEl = null;
    /**
     * A singleton div element to hold the content of a preview. Contained by
     * boxEl.
     * @type {?HTMLElement}
     */
    this.contentEl = null;
  }

  /** Creates the preview div element if it doesn't yet exist. */
  init() {
    if (this.boxEl) {
      return;
    }

    this.boxEl = document.createElement('div');
    this.boxEl.id = 'preview-box';
    this.boxEl.addEventListener('mouseover', (ev) => this.onPreviewMouseOver(ev));
    this.boxEl.addEventListener('mouseout', (ev) => this.onPreviewMouseOut(ev));
    this.boxEl.classList.add('preview-disabled');

    // Div for better box-shadow performance by leveraging GPU accelerated
    // opacity: https://tobiasahlin.com/blog/how-to-animate-box-shadow/.
    const shadow = document.createElement('div');
    shadow.id = 'preview-shadow'

    // Div to hold the preview content.
    this.contentEl = document.createElement('div')
    this.contentEl.id = 'preview-content'

    this.boxEl.appendChild(this.contentEl);
    this.boxEl.appendChild(shadow)
    document.body.append(this.boxEl);
  }

  /** Add event listeners to all preview targets in the document. */
  addListeners() {
    const targets = document.getElementsByClassName('preview-target');
    for (const target of targets) {
      target.addEventListener('mouseover', (ev) => this.onTargetMouseOver(ev));
      target.addEventListener('mouseout', (ev) => this.onTargetMouseOut(ev));
    }
  }

  /**
   * Callback for when the mouse enters the preview target bounding box.
   * @param {Event} ev
   * @return void
   */
  onTargetMouseOver(ev) {
    ev.preventDefault();
    this.init();
    const currentEl = /** @type {Element} */ ev.target;
    const targetEl = currentEl.closest('.preview-target');
    if (!targetEl) {
      console.warn(`preview-box: no surrounding <a> element for ${ev.target}`)
      return
    }

    if (this.currentTarget === targetEl) {
      // We're showing a preview box and the user moved the mouse out and then
      // back-in before the hide timer finished. Keep showing the preview.
      clearTimeout(this.hidePreviewTimer);
    } else {
      // Only request to show preview box if it's not currently displayed to
      // avoid a flicker because we hide the preview box for 1 frame to get the
      // correct height.
      this.showPreviewTimer = setTimeout(
          () => requestAnimationFrame(() => this.showPreviewBox(targetEl)),
          PreviewLifecycle.showPreviewDelayMs);
    }
  }

  /**
   * Callback for when the mouse exits the preview target bounding box.
   * @param {Event} ev
   * @return void
   */
  onTargetMouseOut(ev) {
    ev.preventDefault();
    clearTimeout(this.showPreviewTimer);
    clearTimeout(this.hidePreviewTimer);
    this.hidePreviewTimer = setTimeout(
        () => requestAnimationFrame(() => this.hidePreviewBox()),
        PreviewLifecycle.hidePreviewDelayMs);
  }

  /**
   * Callback for when the mouse enters the preview target bounding box.
   * @param {Event} ev
   * @return void
   */
  onPreviewMouseOver(ev) {
    ev.preventDefault();
    // We moved out of the preview back into to the preview so the user wants to
    // keep using the preview.
    clearTimeout(this.hidePreviewTimer);
  }

  /**
   * Callback for when the mouse exits the preview target bounding box.
   * @param {Event} ev
   * @return void
   */
  onPreviewMouseOut(ev) {
    ev.preventDefault();
    clearTimeout(this.hidePreviewTimer);
    this.hidePreviewTimer = setTimeout(
        () => requestAnimationFrame(() => this.hidePreviewBox()),
        PreviewLifecycle.hidePreviewDelayMs);
  }

  /** Hides the preview box. */
  hidePreviewBox() {
    this.currentTarget = null;
    this.boxEl.classList.add('preview-disabled');
  }

  /**
   * Calls fn on root and each descendant of root using depth-first search.
   * @param {!Element} root
   * @param {!function(Element)} fn
   */
  recurseChildren(root, fn) {
    fn(root)
    for (const child of root.children) {
      this.recurseChildren(child, fn)
    }
  }

  /**
   * Builds content to show in the preview box with info about the target
   * element.
   * @param {HTMLElement} targetEl
   * @return {string} HTML contents of the preview box, or empty if failed to
   * build.
   */
  buildPreviewContent(targetEl) {
    const title = targetEl.dataset.previewTitle;
    const snippet = targetEl.dataset.previewSnippet;
    if (title && snippet) {
      return title + snippet;
    }

    const type = targetEl.dataset.linkType;
    switch (type) {
      case "citation":
        const refId = targetEl.getAttribute('href').slice(1);
        const ref = document.getElementById(refId);
        if (!ref) {
          console.warn(`preview-box: no cite reference found for id='${refId}'`);
          return ""
        }
        // Clone node for easier manipulation.
        const clone = ref.cloneNode(/* deep */ true);
        // Drop the citation number, e.g. "[1]".
        clone.removeChild(clone.childNodes[0]);
        return clone.innerHTML;

      case "cite-reference-num":
        const strIds = targetEl.dataset.citeIds;
        const ids = strIds.split(' ');
        if (ids.length === 0) {
          console.warn(`preview-box: no citation IDs exist for reference='${targetEl.textContent}'`);
          return ""
        }
        const /** Array<HTMLElement> */ refs =  [];
        for (const id of ids) {
          const ref = document.getElementById(id);
          if (!ref) {
            console.warn(`preview-box: no citation found for id='${id}'`);
            continue;
          }
          refs.push(ref)
        }

        const backLinks = [`<ul class="cite-backlinks">`];
        for (const ref of refs) {
          const p1 = ref.parentElement; // Get to <a> containing the <cite>.
          if (!p1) {
            console.warn(`preview-box: no parent for citation id='${ref.id}'`);
            continue;
          }
          const p2 = p1.parentElement; // Get to enclosing elem for <a>.
          if (!p2) {
            console.warn(`preview-box: no grandparent for citation id='${ref.id}'`);
            continue;
          }
          const  clone = /** @type {Element} */ p2.cloneNode(/* deep */ true);
          // Remove ID attributes and highlight the node.
          this.recurseChildren(clone, (e) => {
            if (e.id === ref.id) {
              e.classList.add('cite-backlink-target')
            }
            e.classList.remove('preview-target'); // avoid nested previews
            e.removeAttribute('id'); // avoid duplicate IDs
          });

          backLinks.push(`
            <li>
              <a href="#${ref.id}" class=cite-backlink-back><em>back</em></a>
              <div class="cite-backlink-preview">${clone.innerHTML}</div>
            </li>`)
        }
        backLinks.push(`<ul>`);
        const title = `<p class=preview-title>Citations for this reference</p>`;
        return title + backLinks.join('')

      default:
        console.warn("preview-box: unknown link type: " + type);
    }

    console.warn('preview-box: failed to build content', targetEl);
    return "";
  }

  /**
   * Shows the preview box with content from the data attributes of the target
   * element.
   * @param {HTMLElement} targetEl
   * @return void
   */
  showPreviewBox(targetEl) {
    const content = this.buildPreviewContent(targetEl);
    if (content === "") {
      return;
    }

    this.boxEl.classList.add('preview-disabled');
    // Remove all children to replace them with new title and snippet.
    while (this.contentEl.firstChild) {
      this.contentEl.firstChild.remove()
    }
    this.contentEl.insertAdjacentHTML('afterbegin', content)
    this.contentEl.style.overflowY = null;
    this.contentEl.style.maxHeight = null;
    // Reset transforms so we don't have to correct them in next frame.
    this.boxEl.style.transform = 'translateX(0) translateY(0)';
    this.currentTarget = targetEl;

    // Use another frame because we need the height of the preview box with the
    // HTML content to correctly position it above or below the preview target.
    requestAnimationFrame(() => {
      this.currentTarget = targetEl;
      const targetBox = targetEl.getBoundingClientRect();
      const previewBox = this.boxEl.getBoundingClientRect();

      const horizDelta = this.calcHorizDelta(targetBox, previewBox)

      const {vertDelta, maxHeight, hasScroll} = this.calcVertDelta(
          targetBox, previewBox)

      if (hasScroll) {
        this.contentEl.style.overflowY = 'scroll';
        this.contentEl.style.maxHeight = `${maxHeight}px`;
      }

      this.boxEl.style.transform = `translateX(${horizDelta}px) `
          + `translateY(${vertDelta}px)`;
      this.boxEl.classList.remove('preview-disabled');
    });
  }

  /**
   * Calculates the horizontal delta needed to align the preview box with the
   * target.
   * @param targetBox {DOMRect}
   * @param previewBox {DOMRect}
   * @returns {number}
   */
  calcHorizDelta(targetBox, previewBox) {
    const tb = targetBox;
    const pb = previewBox;
    const docWidth = document.documentElement.clientWidth;
    const marginHoriz = 10; // Breathing room to left and right.

    let horizDelta = tb.right - pb.left;
    // Check if we extend past the viewport and shift left appropriately.
    const hiddenRight = tb.right + pb.width + marginHoriz - docWidth;
    if (hiddenRight > 0) {
      return horizDelta - hiddenRight;
    }

    // If we don't extend past the right edge of the view port, we're
    // aligned with the right edge of the target. Nudge the preview to the
    // left to make it clear that the preview is a child of the target.
    const horizNudge = 20;
    // Don't nudge more than halfway past the element.
    return horizDelta - Math.min(horizNudge, tb.width / 2);
  }

  /**
   * Calculate the vertical delta needed to align the preview box with the
   * target. Also returns the max height and if preview elements needs a scroll
   * bar
   * @param targetBox {DOMRect}
   * @param previewBox {DOMRect}
   * @returns {{hasScroll: boolean, maxHeight: number, vertDelta: number}}
   */
  calcVertDelta(targetBox, previewBox) {
    const tb = targetBox;
    const pb = previewBox;
    const spaceAbove = tb.top;
    const docHeight = document.documentElement.clientHeight;
    const spaceBelow = docHeight - tb.bottom;
    const marginVert = 20; // Breathing room to the top and bottom.

    // Place preview above target by default to avoid masking text below.
    let vertDelta = tb.top - pb.top - pb.height;
    const vertNudge = 4; // Give a little nudge for breathing space.
    let maxHeight = spaceAbove - vertNudge - marginVert;

    if (spaceAbove < pb.height && pb.height < spaceBelow) {
      // Place preview below target only if it can contain the entire preview
      // and the space above cannot.
      console.debug("preview: placing below target - no overflow");
      vertDelta = tb.bottom - pb.top + vertNudge;
      return {vertDelta: vertDelta, maxHeight, hasScroll: false};
    }

    // Otherwise, we're placing below.
    vertDelta -= vertNudge;

    const vertHidden = pb.height - maxHeight;
    if (vertHidden <= 0) {
      console.debug("preview: placing above target - no overflow");
      return {vertDelta, maxHeight, hasScroll: false};
    }

    // The preview extends past the top of the view port.
    console.debug(
        `preview: extends past top of viewport by ${vertHidden}px.`);
    const maxSteal = marginVert * 0.6 + vertNudge * 0.6;
    // Remove the scrollbar by stealing padding.
    if (vertHidden < maxSteal) {
      console.debug('preview: avoiding scrollbar by stealing padding');
      return {
        vertDelta: vertDelta - vertHidden,
        maxHeight: maxHeight + vertHidden,
        hasScroll: false
      };
    }

    console.debug('preview: using vertical scroll bar')
    this.contentEl.style.overflowY = 'scroll';
    return {vertDelta: vertDelta + vertHidden, maxHeight, hasScroll: true}
  }
}

PreviewLifecycle.showPreviewDelayMs = 300;
// Hiding feels better if a bit faster. Usually you want to hide things
// "instantly."
PreviewLifecycle.hidePreviewDelayMs = 200;

// Preview hovers.
// Each preview target contains data attributes describing how to display
// information about the target. The attributes include:
// - data-title: required, the title of the link.
// - data-snippet: required, a short snippet about the link.
// On hover, we re-use a global element, #preview-box, to display the
// attributes. The preview is a no-op on devices with touch.
(() => {
  // Detect touch based devices as a proxy for not having hover.
  // https://stackoverflow.com/a/8758536/30900
  let hasHover = false;
  try {
    document.createEvent("TouchEvent");
  } catch (e) {
    hasHover = true;
  }
  if (!hasHover) {
    console.debug("preview: no hover support, skipping previews")
    return;
  }

  console.debug("preview: hover supported, enabling previews")
  const preview = new PreviewLifecycle();
  preview.addListeners();
})();
