const { defaults: tsjPreset } = require('ts-jest/presets');

module.exports = {
  errorOnDeprecated: true,
  clearMocks: true,
  transform: {
    ...tsjPreset.transform,
  },
  setupFilesAfterEnv: ['./src/testing/global_jest_setup'],
  testEnvironment: 'node',
  testMatch: ["**/*_test.js", "**/*_test.ts"]
};
