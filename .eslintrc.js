module.exports =  {
  parser:  '@typescript-eslint/parser',
  parserOptions: {
    "project": "./tsconfig.json"
  },
  extends:  [
    'plugin:@typescript-eslint/recommended',
    "plugin:@typescript-eslint/recommended-requiring-type-checking",
    'prettier/@typescript-eslint',
  ],
  rules:  {
    // Having to move functions around to work-around hoisting is annoying.
    '@typescript-eslint/no-use-before-define': 'off',
    // Inferrable types don't hurt code readability. Too pedantic.
    '@typescript-eslint/no-inferrable-types': 'off',
    // Allow a single extends.
    "@typescript-eslint/no-empty-interface": [
      "error", { "allowSingleExtends": true }
    ],
    // We want to allow private constructors.
    // https://github.com/typescript-eslint/typescript-eslint/issues/1178
    '@typescript-eslint/no-empty-function': 'off',

    // False positives for imported types.
    // https://github.com/typescript-eslint/typescript-eslint/issues/363
    '@typescript-eslint/no-unused-vars': 'off',

    // Allow using async functions in callbacks. Otherwise, this will error
    // for all async functions because they return Promise<void>.
    // fs.readFile('file.txt', async (err, txt) => {
    //   await doThing(err, txt);
    // }
    "@typescript-eslint/no-misused-promises": [
      "error", { checksVoidReturn: false }
    ],

    // Functions as expressions are usually simple enough not to need types.
    "@typescript-eslint/explicit-function-return-type": ["warn", {
      allowExpressions: true,
    }],

    // Floating promises should always be handled explicitly.
    "@typescript-eslint/no-floating-promises": ["warn"],
  },
};
