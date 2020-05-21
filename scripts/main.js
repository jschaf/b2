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
      return;
    }
    await new Promise(resolve => setTimeout(resolve, intervalMs));
  }
  window.adblockStatus = 'none';
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
 * The complexity comes from interactions between the target link and preview
 * div. Use-cases we want to support:
 *
 * - Allow grace period to continue showing the preview when moving from the
 *   target to preview box.
 *
 * - Allow grace period to continue showing the preview when leaving the box
 *   but quickly returning.
 */
class PreviewLifecycle {
  constructor() {
    /**
     * The current, displayed preview target pending or displayed. If no preview
     * is displayed, currentTarget is null.
     * @type {?HTMLElement}
     */
    this.currentTarget = null;
    /** @type {?TimeoutId} */
    this.showPreviewTimer = null;
    /** @type {?TimeoutId} */
    this.hidePreviewTimer = null;
    /**
     * A singleton div element to hold previews of preview target links.
     * Lazily initialized on the first hover of a preview target.
     * @type {?HTMLElement}
     */
    this.previewEl = null;
  }

  /** Creates the preview div element if it doesn't yet exist. */
  init() {
    if (this.previewEl) {
      return;
    }

    const el = this.previewEl = document.createElement('div');
    el.id = 'preview-box';
    // Use visibility instead of display: none so that the position is accurate.
    el.addEventListener('mouseover', (ev) => this.onPreviewMouseOver(ev));
    el.addEventListener('mouseout', (ev) => this.onPreviewMouseOut(ev));
    // Create another element for better box-shadow performance.
    // https://tobiasahlin.com/blog/how-to-animate-box-shadow/
    const shadow = document.createElement('div');
    shadow.id = 'preview-shadow'
    el.appendChild(shadow);
    document.body.append(el);
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
    const targetEl = ev.target.closest('.preview-target');
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
    this.previewEl.classList.add('preview-disabled');
  }

  /**
   * Shows the preview box with content from the data attributes of the target
   * element.
   * @param {HTMLElement} targetEl
   * @return void
   */
  showPreviewBox(targetEl) {
    const title = targetEl.dataset.previewTitle;
    const snippet = targetEl.dataset.previewSnippet;
    if (!title || !snippet) {
      console.warn('preview-box: missing data-title or data-snippet attrs',
          targetEl)
      return;
    }
    this.currentTarget = targetEl;

    const docWidth = document.documentElement.clientWidth;
    const docHeight = document.documentElement.clientHeight;
    const marginHoriz = 10; // Breathing room to left and right.

    this.previewEl.classList.add('preview-disabled');
    // Remove all children except the last, the preview shadow.
    while (this.previewEl.childNodes.length > 1) {
      this.previewEl.removeChild(this.previewEl.firstChild);
    }
    this.previewEl.insertAdjacentHTML('afterbegin', title + snippet)
    const previewWidth = Math.min(580, docWidth - 2 * marginHoriz)
    this.previewEl.style.width = `${previewWidth}px`;
    // Reset transforms so we don't have to correct them in next frame.
    this.previewEl.style.transform = 'translateX(0) translateY(0)';

    // Use another frame because we need the height of the preview box with the
    // HTML content to correctly position it above or below the preview target.
    requestAnimationFrame(() => {
      this.currentTarget = targetEl;
      const t = targetEl.getBoundingClientRect();
      const p = this.previewEl.getBoundingClientRect();
      const spaceAbove = t.top;
      const spaceBelow = docHeight - t.bottom;

      let diffLeft = t.right - p.left;
      // Check if we extend past the viewport and shift left appropriately.
      const hiddenRight = t.right + p.width + marginHoriz - docWidth;
      if (hiddenRight > 0) {
        diffLeft -= hiddenRight;
      } else {
        // If we don't extend past the right edge of the view port, we're
        // aligned with the right edge of the target. Nudge the preview to the
        // left to make it clear that the preview is a child of the target.
        const nudge = 20;
        diffLeft -= Math.min(nudge, t.width / 2);
      }
      // Place preview above target by default to avoid masking text below.
      let diffTop = t.top - p.top - this.previewEl.offsetHeight;
      const vertNudge = 4; // Give a little nudge for breathing space.
      if (p.height > spaceAbove && p.height < spaceBelow) {
        // Place preview below target only if it can contain the entire preview
        // and the space above cannot.
        diffTop = t.bottom - p.top + vertNudge;
      } else {
        diffTop -= vertNudge;
      }

      this.previewEl.style.transform = `translateX(${diffLeft}px) translateY(${diffTop}px)`;
      this.previewEl.classList.remove('preview-disabled');
    });
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
    return;
  }

  const preview = new PreviewLifecycle();
  preview.addListeners();
})();
