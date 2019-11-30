// https://jestjs.io/docs/en/configuration.html#snapshotresolver-string
module.exports = {
  /** Resolves from test to snapshot path. */
  resolveSnapshotPath: (testPath, _snapshotExtension) => {
    return testPath.replace('_test', '_test_snap');
  },

  /** Resolves from snapshot to test path. */
  resolveTestPath: (snapshotFilePath, _snapshotExtension) => {
    return snapshotFilePath.replace('_snap', '');
  },

  /**
   * Example test path, used for preflight consistency check of the
   * implementation above.
   */
  testPathForConsistencyCheck: 'some/example_test.ts',
};
