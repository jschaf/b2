const OFF = 'off';
const ERROR = 'error';
const WARN = 'warn';

module.exports = {
  parser: '@typescript-eslint/parser',
  parserOptions: {
    project: './tsconfig.json',
  },
  extends: [
    'plugin:@typescript-eslint/recommended',
    'plugin:@typescript-eslint/recommended-requiring-type-checking',
    'prettier/@typescript-eslint',
  ],
  rules: {
    // Having to move functions around to work-around hoisting is annoying.
    '@typescript-eslint/no-use-before-define': OFF,
    // Inferrable types don't hurt code readability. Too pedantic.
    '@typescript-eslint/no-inferrable-types': OFF,
    // Allow a single extends.
    '@typescript-eslint/no-empty-interface': [
      ERROR,
      { allowSingleExtends: true },
    ],
    // We want to allow private constructors.
    // https://github.com/typescript-eslint/typescript-eslint/issues/1178
    '@typescript-eslint/no-empty-function': OFF,

    // False positives for imported types.
    // https://github.com/typescript-eslint/typescript-eslint/issues/363
    '@typescript-eslint/no-unused-vars': OFF,

    // Allow using async functions in callbacks. Otherwise, this will error
    // for all async functions because they return Promise<void>.
    // fs.readFile('file.txt', async (err, txt) => {
    //   await doThing(err, txt);
    // }
    '@typescript-eslint/no-misused-promises': [
      ERROR,
      { checksVoidReturn: false },
    ],

    // Functions as expressions are usually simple enough not to need types.
    '@typescript-eslint/explicit-function-return-type': [
      WARN,
      {
        allowExpressions: true,
      },
    ],

    // Floating promises should always be handled explicitly.
    '@typescript-eslint/no-floating-promises': [ERROR],
  },
  overrides: [
    {
      files: ['**/*.js'],
      rules: {
        // Allow require in js files.
        '@typescript-eslint/no-var-requires': OFF,
      },
    },
  ],
};
