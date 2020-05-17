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

/** Utilities for promises. */
const promises = {
  /** Returns a promise that resolves after the given duration in milliseconds. */
  sleep(durationMs) {
    return new Promise((resolve) => {
      setTimeout(() => resolve(), durationMs);
    });
  }
}

/**
 * @param {HTMLElement} target
 * @return {!{top: number, left: number, right: number, bottom: number}}
 */
const getBounds = (target) => {
  const targetBox = target.getBoundingClientRect();
  return {
    top: targetBox.top - document.documentElement.clientTop,
    left: targetBox.left - document.documentElement.clientLeft,
    right: document.body.clientWidth - targetBox.width - targetBox.left,
    bottom: document.body.clientHeight - targetBox.height - targetBox.top,
  }
}

/**
 * PreviewLifecycle manages the state transitions for the preview box display.
 */
class PreviewLifecycle {
  constructor() {
    /**
     * The current preview target pending or displayed. If no preview is
     * displayed and no preview is pending, currentTarget is null.
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
    el.style.visibility = 'hidden';
    el.addEventListener('mouseover', (ev) => this.onPreviewMouseOver(ev));
    el.addEventListener('mouseout', (ev) => this.onPreviewMouseOut(ev));
    document.body.append(el);
  }

  /** Add event listeners to all preview targets in the document. */
  addListeners() {
    const previewTargetClass = 'preview-target'
    const targets = document.getElementsByClassName(previewTargetClass);
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
    const targetEl = ev.target.closest('a');
    if (!targetEl) {
      console.warn(`preview-box: no surrounding <a> element for ${ev.target}`)
      return
    }

    // We're showing a preview box and the user moved the mouse out and then
    // back-in before the hide timer finished. Keep showing the preview.
    if (this.currentTarget === targetEl) {
      clearTimeout(this.hidePreviewTimer);
    }

    this.showPreviewTimer = setTimeout(
        () => requestAnimationFrame(() => this.showPreviewBox(targetEl)),
        PreviewLifecycle.showPreviewDelayMs);
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
    // We moved from the target to the preview so the user wants to keep using
    // the preview.
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

  hidePreviewBox() {
    if (this.previewEl.style.visibility !== 'hidden') {
      this.previewEl.style.visibility = 'hidden';
    }
  }

  /**
   * @param {HTMLElement} targetEl
   * @return void
   */
  showPreviewBox(targetEl) {
    const title = targetEl.dataset.title;
    const snippet = targetEl.dataset.snippet;
    if (!title || !snippet) {
      console.warn('preview-box: missing data-title or data-snippet attrs',
          targetEl)
      return;
    }
    this.currentTarget = targetEl;

    const previewHTML = `<h3>${title}</h3><p>${snippet}</p>`;

    // Update the transform CSS.
    const {left: tLeft, top: tTop} = getBounds(targetEl);
    const {left: pLeft, top: pTop} = getBounds(this.previewEl);
    // Example transform string: 'translateX(653.484px) translateY(151.062px)'
    const oldTransform = this.previewEl.style.transform;
    const [_, xOffs, yOffs] = oldTransform.match(
        /^translateX\(([-0-9.]+)px\) translateY\(([-0-9.]+)px\)/)
    || ['0', '0', '0'];
    const diffLeft = tLeft - pLeft + (+xOffs);
    const diffTop = tTop - pTop + (+yOffs) - 100;
    const newTransform = `translateX(${diffLeft}px) translateY(${diffTop}px)`;

    // Avoid changing inner HTML if no change.
    if (this.previewEl.innerHTML !== previewHTML) {
      this.previewEl.innerHTML = previewHTML;
    }
    // CSS changes trigger a repaint. Avoid if nothing changed.
    if (newTransform !== oldTransform) {
      this.previewEl.style.transform = newTransform;
    }
    if (this.previewEl.style.visibility !== 'visible') {
      this.previewEl.style.visibility = 'visible';
    }
  }
}

PreviewLifecycle.showPreviewDelayMs = 300;
PreviewLifecycle.hidePreviewDelayMs = 300;

// Preview hovers.
// Each preview target contains data attributes describing how to display
// information about the target. The attributes include:
// - data-title: required, the title of the link.
// - data-snippet: required, a short snippet about the link.
// On hover, we re-use a global element, #preview-box, to display the
// attributes. The preview is a no-op on devices with touch.
(async () => {
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
