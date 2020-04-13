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
