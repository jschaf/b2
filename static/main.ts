type AdblockStatus = 'unknown' | 'blocking' | 'none';

declare interface Window {
  adblockStatus: AdblockStatus;
  heap: HeapAnalytics;
}

declare interface Navigator {
  connection?: NetworkInformation;
}

declare interface NetworkInformation {
  saveData?: boolean;
  effectiveType?: 'slow-2g'| '2g'| '3g'| '4g';
}

type EventProps = Record<string, string | number>;

interface HeapCfg {
  disableTextCapture?: boolean,
  secureCookie?: boolean,
  trackingServer?: string,
}

enum TrackName {
  HoverPreview = 'hover preview'
}

class HeapAnalytics extends Array<any> {
  userId?: string;
  identity?: string;

  private constructor(readonly appid: string, readonly config: HeapCfg) {
    super();
  }

  static forEnvId(envId: string, config: HeapCfg): HeapAnalytics {
    return new HeapAnalytics(envId, config);
  }

  track(name: TrackName, props: EventProps): void {
    this.push(['track', name, props]);
  }

  identify(id: string): void {
    this.push(['identify', id]);
  }

  resetIdentity(): void {
    this.push(['resetIdentity']);
  }

  addUserProperties(props: EventProps): void {
    this.push(['addUserProperties', props]);
  }

  addEventProperties(props: EventProps): void {
    this.push(['addEventProperties', props]);
  }

  removeEventProperty(prop: string): void {
    this.push(['removeEventProperty', prop]);
  }

  clearEventProperties(): void {
    this.push(['clearEventProperties']);
  }
}

const checkDef = <T>(x: T, msg?: string): NonNullable<T> => {
  if (x === null) {
    throw new Error(msg ?? `Expression was null but expected non-null.`);
  }
  if (x === undefined) {
    throw new Error(msg ?? `Expression was undefined but expected non-null.`);
  }
  return x as NonNullable<T>;
};

const checkHtmlEL = (x: unknown, msg?: string): HTMLElement => {
  if (x instanceof HTMLElement) {
    return x as HTMLElement;
  }
  throw new Error(msg ?? `Expected element to be an HTMLElement but was ${JSON.stringify(x)}.`)
};

function assertDef<T>(x: T, msg?: string): asserts x is NonNullable<T> {
  checkDef(x, msg);
}

const checkInstance = <T extends any>(x: unknown, typ: T, msg?: string): NonNullable<T> => {
  checkDef(x, 'Express must be non-null for instanceof check');
  if (!(x instanceof (typ as any))) {
    throw new Error(msg ?? `Expression had instanceof ${x} but expected ${typ}`);
  }
  return x as NonNullable<T>;
};

/** A simple logger. */
class Logger {
  private constructor() {
  }

  static forConsole(): Logger {
    return new Logger();
  }

  warn(...data: any[]) {
    console.log(...data);
  }

  info(...data: any[]) {
    console.log(...data);
  }

  debug(...data: any[]) {
    console.debug(...data);
  }
}

const log = Logger.forConsole();

// Creates the heap stub that records API calls to eventually replay while the
// real heap.js downloads. The real heap.js is templated into base.gohtml.
window.heap = HeapAnalytics.forEnvId(
    '1506018335',
    {trackingServer: 'https://joe.schafer.dev'}
);

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
      log.debug(`adblock status: ${window.adblockStatus}`);
      return;
    }
    await new Promise(resolve => setTimeout(resolve, intervalMs));
  }
  window.adblockStatus = 'none';
  window.heap.addEventProperties({ adblock: window.adblockStatus });
  log.debug(`adblock status: ${window.adblockStatus}`);
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
  static readonly showPreviewDelayMs = 300;
  // Hiding feels better if a bit faster. Usually you want to hide things
  // "instantly."
  static readonly hidePreviewDelayMs = 200;

  // The current, displayed preview target pending or displayed. If no preview
  // is displayed, currentTarget is null.
  private currentTarget: HTMLLinkElement | null = null;

  private hoverStart: number | null = null;
  private showPreviewTimer: number = 0;
  private hidePreviewTimer: number = 0;
  // A singleton div element to hold previews of preview target links.
  // Lazily initialized on the first hover of a preview target. Contains
  // contentEl.
  private boxEl: HTMLElement | null = null;
  // A singleton div element to hold the content of a preview. Contained by
  // boxEl.
  private contentEl: HTMLElement | null = null;

  constructor() {
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
    shadow.id = 'preview-shadow';

    // Div to hold the preview content.
    this.contentEl = document.createElement('div');
    this.contentEl.id = 'preview-content';

    this.boxEl.appendChild(this.contentEl);
    this.boxEl.appendChild(shadow);
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

  /** Callback for when the mouse enters the preview target bounding box. */
  onTargetMouseOver(ev: Event): void {
    ev.preventDefault();
    this.init();
    const currentEl = checkDef(ev.target, 'preview target mouse over') as HTMLElement;
    const targetEl = currentEl.closest('.preview-target') as HTMLLinkElement;
    if (!targetEl) {
      log.info(`preview-box: no surrounding <a> element for ${ev.target}`);
      return;
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

  /** Callback for when the mouse exits the preview target bounding box. */
  onTargetMouseOut(ev: Event): void {
    ev.preventDefault();
    clearTimeout(this.showPreviewTimer);
    clearTimeout(this.hidePreviewTimer);
    this.hidePreviewTimer = setTimeout(
        () => requestAnimationFrame(() => this.hidePreviewBox()),
        PreviewLifecycle.hidePreviewDelayMs);
  }

  /** Callback for when the mouse enters the preview target bounding box. */
  onPreviewMouseOver(ev: Event): void {
    ev.preventDefault();
    // We moved out of the preview back into to the preview so the user wants to
    // keep using the preview.
    clearTimeout(this.hidePreviewTimer);
  }

  /** Callback for when the mouse exits the preview target bounding box. */
  onPreviewMouseOut(ev: Event): void {
    ev.preventDefault();
    clearTimeout(this.hidePreviewTimer);
    this.hidePreviewTimer = setTimeout(
        () => requestAnimationFrame(() => this.hidePreviewBox()),
        PreviewLifecycle.hidePreviewDelayMs);
  }

  /** Hides the preview box. */
  hidePreviewBox() {
    if (this.currentTarget != null) {
      window.heap.track(TrackName.HoverPreview, {
        hoverTarget: checkDef(this.currentTarget).href,
        durationMillis: this.hoverStart ? Date.now() - this.hoverStart : 0,
      });
    }
    this.currentTarget = null;
    this.hoverStart = null;
    if (!this.boxEl) {
      log.warn(`preview-box: boxEl was null but called hidePreviewBox`);
      return;
    }
    this.boxEl.classList.add('preview-disabled');
  }

  /** Calls fn on root and each descendant of root using depth-first search. */
  recurseChildren(root: Element, fn: (e: Element) => void) {
    fn(root);
    for (const child of root.children) {
      this.recurseChildren(child, fn);
    }
  }

  /**
   * Builds content to show in the preview box with info about the target
   * element. Returns the HTML contents of the preview box, or empty if failed
   * to build the preview box.
   */
  buildPreviewContent(targetEl: HTMLElement): string {
    const title = targetEl.dataset.previewTitle ?? '';
    const snippet = targetEl.dataset.previewSnippet;
    if (snippet) {
      return title + snippet;
    }

    const type = targetEl.dataset.linkType;
    switch (type) {
      case 'cite-reference-num':
        const strIds = checkDef(targetEl.dataset.citeIds, `no citeIds attribute found`);
        const ids = strIds.split(' ');
        if (ids.length === 0) {
          log.warn(`preview-box: no citation IDs exist for reference='${targetEl.textContent}'`);
          return '';
        }
        const refs: HTMLElement[] = [];
        for (const id of ids) {
          const ref = document.getElementById(id);
          if (!ref) {
            log.warn(`preview-box: no citation found for id='${id}'`);
            continue;
          }
          refs.push(ref);
        }

        const backLinks = [`<ul class="cite-backlinks">`];
        for (const ref of refs) {
          const p1 = ref.parentElement; // Get to <a> containing the <cite>.
          if (!p1) {
            log.warn(`preview-box: no parent for citation id='${ref.id}'`);
            continue;
          }
          const p2 = p1.parentElement; // Get to enclosing elem for <a>.
          if (!p2) {
            log.warn(`preview-box: no grandparent for citation id='${ref.id}'`);
            continue;
          }
          const clone = p2.cloneNode(true) as HTMLElement;
          // Remove ID attributes and highlight the node.
          this.recurseChildren(clone, (e) => {
            if (e.id === ref.id) {
              e.classList.add('cite-backlink-target');
            }
            e.classList.remove('preview-target'); // avoid nested previews
            e.removeAttribute('id'); // avoid duplicate IDs
          });

          backLinks.push(`
            <li>
              <div class="cite-backlink-preview">
                <a href="#${ref.id}" class=cite-backlink-back><em>back</em></a>
                ${clone.innerHTML}
              </div>
            </li>`);
        }
        backLinks.push(`<ul>`);
        const title = `<p class=preview-title>Citations for this reference</p>`;
        return title + backLinks.join('');

      default:
        log.warn('preview-box: unknown link type: ' + type);
    }

    log.warn('preview-box: failed to build content', targetEl);
    return '';
  }

  /**
   * Shows the preview box with content from the data attributes of the target
   * element.
   */
  showPreviewBox(targetEl: HTMLLinkElement): void {
    const content = this.buildPreviewContent(targetEl);
    if (content === '') {
      return;
    }
    assertDef(this.boxEl, `boxEl was null for showPreviewBox`);
    assertDef(this.contentEl, `contentEl was null for showPreviewBox`);

    this.boxEl.classList.add('preview-disabled');
    // Remove all children to replace them with new title and snippet.
    while (this.contentEl.firstChild) {
      this.contentEl.firstChild.remove();
    }
    this.contentEl.insertAdjacentHTML('afterbegin', content);
    this.contentEl.style.overflowY = '';
    this.contentEl.style.maxHeight = '';
    // Reset transforms so we don't have to correct them in next frame.
    this.boxEl.style.transform = 'translateX(0) translateY(0)';
    this.currentTarget = targetEl;
    this.hoverStart = Date.now();

    // Use another frame because we need the height of the preview box with the
    // HTML content to correctly position it above or below the preview target.
    requestAnimationFrame(() => {
      assertDef(this.boxEl, `boxEl was null for showPreviewBox requestAnimationFrame`);
      assertDef(this.contentEl, `contentEl was null for showPreviewBox requestAnimationFrame`);
      this.currentTarget = targetEl;
      const targetBox = targetEl.getBoundingClientRect();
      const previewBox = this.boxEl.getBoundingClientRect();

      const horizDelta = this.calcHorizDelta(targetBox, previewBox);

      const { vertDelta, maxHeight, hasScroll } = this.calcVertDelta(
          targetBox, previewBox);

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
   */
  calcHorizDelta(targetBox: DOMRect, previewBox: DOMRect): number {
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
   */
  calcVertDelta(targetBox: DOMRect, previewBox: DOMRect): { hasScroll: boolean, maxHeight: number, vertDelta: number } {
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
      log.debug('preview: placing below target - no overflow');
      vertDelta = tb.bottom - pb.top + vertNudge;
      return { vertDelta: vertDelta, maxHeight, hasScroll: false };
    }

    // Otherwise, we're placing below.
    vertDelta -= vertNudge;

    const vertHidden = pb.height - maxHeight;
    if (vertHidden <= 0) {
      log.debug('preview: placing above target - no overflow');
      return { vertDelta, maxHeight, hasScroll: false };
    }

    // The preview extends past the top of the view port.
    log.debug(
        `preview: extends past top of viewport by ${vertHidden}px.`);
    const maxSteal = marginVert * 0.6 + vertNudge * 0.6;
    // Remove the scrollbar by stealing padding.
    if (vertHidden < maxSteal) {
      log.debug('preview: avoiding scrollbar by stealing padding');
      return {
        vertDelta: vertDelta - vertHidden,
        maxHeight: maxHeight + vertHidden,
        hasScroll: false,
      };
    }

    log.debug('preview: using vertical scroll bar');
    assertDef(this.contentEl, `contentEl was null for calcVertDelta`);
    this.contentEl.style.overflowY = 'scroll';
    return { vertDelta: vertDelta + vertHidden, maxHeight, hasScroll: true };
  }
}

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
    document.createEvent('TouchEvent');
  } catch (e) {
    hasHover = true;
  }
  if (!hasHover) {
    log.debug('preview: no hover support, skipping previews');
    return;
  }

  log.debug('preview: hover supported, enabling previews');
  const preview = new PreviewLifecycle();
  preview.addListeners();
})();

// Copy heading link when clicking the paragraph symbol.
(() => {
  const copySourceEl = document.createElement('input');
  copySourceEl.id = 'copy-source';
  copySourceEl.type = 'hidden';
  document.body.append(copySourceEl);

  const copyHeadingUrl = (ev: Event) => {
    ev.preventDefault();
    const paraEl = ev.target as HTMLLinkElement;
    if (paraEl == null) {
      log.warn('copy-heading-url: event target is undefined');
      return;
    }
    const headingHref = paraEl.href;
    if (headingHref == null || headingHref === '') {
      log.warn('copy-heading-url: event target href is not defined');
      return;
    }
    const headingUrl = new URL(headingHref);
    const targetHash = headingUrl.hash;

    const { origin, pathname } = window.location;
    const newUrl = origin + pathname + targetHash;
    history.replaceState(null, '', newUrl);
    copySourceEl.type = ''; // show the input so we can select it
    copySourceEl.value = newUrl;
    copySourceEl.focus({ preventScroll: true });
    document.execCommand('SelectAll');
    document.execCommand('Copy', /* show UI */ false, '');
    copySourceEl.type = 'hidden';
  };

  const targets = document.getElementsByClassName('heading-anchor');
  for (const target of targets) {
    target.addEventListener('click', (ev) => copyHeadingUrl(ev));
  }
})();

// Prefetch URLs on the whitelisted domains on mouseover or touch start events.
// Forked from instant.page, https://instant.page/license.
// TODO: Allow other whitelisted domains.
(() => {
  const preloads = new Set<string>() // hrefs already preloaded
  let mouseoverTimer = 0;
  const prefetcher = document.createElement("link");
  const supportsPrefetch = prefetcher?.relList?.supports("prefetch") ?? false;
  const isSavingData = navigator?.connection?.saveData ?? false;
  const conn = navigator?.connection?.effectiveType ?? 'unknown'
  const is2gConn = conn.includes('2g');
  if (!supportsPrefetch || isSavingData || is2gConn) {
    log.debug(`prefetch: disabled`);
    return;
  }
  log.debug(`prefetch: enabled`);
  prefetcher.rel = "prefetch";
  document.head.appendChild(prefetcher);

  const preload = (url: string): void => {
    prefetcher.rel = 'prefetch';
    prefetcher.href = url;
    preloads.add(url);
  }

  const onMouseout = (ev: MouseEvent): void => {
    // On mouseout, target is the element we exited and relatedTarget is elem
    // we entered (or null).
    let exitLink = checkHtmlEL(ev.target).closest("a");
    let enterLink = (ev?.relatedTarget as HTMLLinkElement)?.closest("a");
    if (ev.relatedTarget && exitLink === enterLink) {
      return;
    }
    if (mouseoverTimer !== 0) {
      log.debug(`prefetch: canceling prefetch ${exitLink?.href}`);
      clearTimeout(mouseoverTimer);
      mouseoverTimer = 0;
    }
  };

  // Returns true if we should preload this anchor link, otherwise false.
  const shouldPreload = (node: HTMLAnchorElement): boolean => {
    if (!node) {
      return false;
    }
    if (node.protocol !== 'http:' && node.protocol !== 'https:') {
      return false;
    }
    if (node.protocol === 'http:' && location.protocol === 'https:') {
      return false;
    }
    if (preloads.has(node.href)) {
      log.debug(`prefetch: skipping url ${node.href}, already preloaded`);
      return false;
    }

    const url = new URL(node.href);
    if (url.origin !== location.origin) {
      log.debug(`prefetch: skipping url ${url}, different origin`);
      return false;
    }

    const dest = url.pathname + url.search;
    const cur = location.pathname + location.search;
    if (dest === cur) {
      log.debug(`prefetch: skipping url ${url}, exactly same as current`);
      return false;
    }
    // Assume URLs that only differ by a trailing slash are the same.
    if (cur.slice(-1) === '/' && dest === cur.slice(0, -1)) {
      log.debug(`prefetch: skipping url ${url}, same with trailing slash as current`);
      return false;
    }
    return true;
  };

  // On touchstart, immediately preload the link.
  document.addEventListener("touchstart", (touchEv) => {
    const link = checkHtmlEL(touchEv.target).closest("a");
    if (!link || !shouldPreload(link)) {
      return;
    }
    preload(link.href)
  }, {capture: true, passive: true});

  // On mouseover, preload the link after a delay.
  document.addEventListener("mouseover", (ev: MouseEvent) => {
    // Browsers emulate mouse events from touch events so mouseover will be
    // called after touchstart. We'll avoid double preloading because
    // shouldPreload checks to see if we've already loaded a URL.
    const link = checkHtmlEL(ev.target).closest("a");
    if (!link || !shouldPreload(link)) {
      return;
    }
    link.addEventListener("mouseout", onMouseout, {passive: true});
    const delayOnHover = 65;
    mouseoverTimer = setTimeout(() => {
      log.debug(`prefetch: loading mouseover link ${link.href}`);
      preload(link.href)
      mouseoverTimer = 0;
    }, delayOnHover);
  }, {capture: true, passive: true});
})();
