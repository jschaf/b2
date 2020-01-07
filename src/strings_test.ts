import { dedent, isOptionalString, isString } from '//strings';

test('isString should work', () => {
  expect(isString(1)).toBe(false);
  expect(isString(false)).toBe(false);
  expect(isString(null)).toBe(false);
  expect(isString(undefined)).toBe(false);
  expect(isString({})).toBe(false);
  expect(isString([])).toBe(false);

  expect(isString('')).toBe(true);
  // noinspection JSPrimitiveTypeWrapperUsage
  expect(isString(new String())).toBe(true);
});

test('isOptionalString should work', () => {
  expect(isOptionalString(1)).toBe(false);
  expect(isOptionalString(false)).toBe(false);
  expect(isOptionalString(null)).toBe(false);
  expect(isOptionalString({})).toBe(false);
  expect(isOptionalString([])).toBe(false);

  expect(isOptionalString(undefined)).toBe(true);
  expect(isOptionalString('')).toBe(true);
  // noinspection JSPrimitiveTypeWrapperUsage
  expect(isOptionalString(new String())).toBe(true);
});

describe('dedent', () => {
  const testData: [string, string, string][] = [
    ['simple strings', 'foo', 'foo'],
    ['leading space 1 line', '   foo', 'foo'],
    ['trailing space 1 line', 'foo  ', 'foo'],
    ['leading space 3 lines', '  foo\n  bar\n  qux\n', 'foo\nbar\nqux'],
    ['trim same space 3 lines', '  foo\n    bar\n   qux\n', 'foo\n  bar\n qux'],
  ];
  for (const [name, input, expected] of testData) {
    it(name, () => {
      const actual = dedent`${input}`;
      expect(actual).toEqual(expected);
    });
  }
});
