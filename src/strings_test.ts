import {isString} from "./strings";

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
