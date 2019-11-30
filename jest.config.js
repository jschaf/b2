const { defaults: tsjPreset } = require('ts-jest/presets');

module.exports = {
  errorOnDeprecated: true,
  clearMocks: true,
  maxConcurrency: 20,
  moduleFileExtensions: ['ts', 'node', 'js', 'json'],
  moduleNameMapper: {
    '//(.*)': '<rootDir>/src/$1',
  },
  transform: {
    ...tsjPreset.transform,
  },
  setupFilesAfterEnv: ['./src/testing/global_jest_setup'],
  testEnvironment: 'node',
  testMatch: ['**/*_test.ts'],
  snapshotResolver: './src/testing/jest_snapshot_resolver',
};
