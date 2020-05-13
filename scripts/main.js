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

  const previewBox = document.createElement('div');
  previewBox.id = 'preview-box';
  document.body.append(previewBox);

  const previewTargetClass = 'preview-target'
  const targets = document.getElementsByClassName(previewTargetClass);
  console.log('!!! targets', targets);

  const onTargetMouseOver = (ev) => {
    ev.preventDefault();
    const target = ev.target.closest('a');
    const {left: pLeft, top: pTop} = previewBox.getBoundingClientRect();
    const {left: tLeft, top: tTop} = target.getBoundingClientRect();

    console.log('mouseover', ev)
  }

  const onTargetMouseOut = (ev) => {
    ev.preventDefault();
    console.log('mouseout', ev)
  }

  for (const target of targets) {
    target.addEventListener('mouseover', onTargetMouseOver);
    target.addEventListener('mouseout', onTargetMouseOut);
  }
})();
