import { dedent, isString } from './strings';

test('isString should work', () => {
  expect(isString(1)).toBe(false);
  expect(isString(false)).toBe(false);
  expect(isString(null)).toBe(false);
  expect(isString(undefined)).toBe(false);
  expect(isString({})).toBe(false);

  expect(isString('')).toBe(true);
  // noinspection JSPrimitiveTypeWrapperUsage
  expect(isString(new String())).toBe(true);
});

describe('dedent', () => {
  test('should work for simple strings', () => {
    expect(dedent`foo`).toEqual('foo');
    expect(dedent`foo bar`).toEqual('foo bar');
  });

  test('should remove leading space from single line', () => {
    expect(dedent`  foo`).toEqual('foo');
  });

  test('should remove leading space from many lines', () => {
    expect(dedent`
      foo
      bar
        qux
      baz  
    `).toEqual('foo\nbar\n  qux\nbaz');
  });

  test('should remove trailing space from single line', () => {
    expect(dedent`  foo   `).toEqual('foo');
  });
});
