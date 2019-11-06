module.exports =  {
  parser:  '@typescript-eslint/parser',
  extends:  [
    'plugin:@typescript-eslint/recommended',
    'prettier/@typescript-eslint',
  ],
  parserOptions:  {
    ecmaVersion:  2018,
    sourceType:  'module',
  },
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

    '@typescript-eslint/explicit-function-return-type': ['warn', {
      allowExpressions: true
    }]
  },
};
