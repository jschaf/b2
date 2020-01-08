/**
 * This package contains custom Jest matchers.
 *
 * See https://stackoverflow.com/a/45745719/30900 for an example of how to write
 * a custom matcher.
 *
 * Update global_jest_matchers.d.ts for Typescript type-checking.
 */

import { isString } from '//strings';
import { Mempost, normalizeHTML, normalizeMempostEntry } from '//post/mempost';

function toEqualMempost(
  this: jest.MatcherContext,
  received: unknown,
  expected: Mempost | Record<string, string>
): jest.CustomMatcherResult {
  if (!(received instanceof Mempost)) {
    return {
      pass: false,
      message: () => `Expected to receive Mempost type but had ${received}`,
    };
  }
  const receivedObj: Record<string, string | Buffer> = {};
  const expectedObj: Record<string, string | Buffer> = {};
  for (const [path, buf] of Object.entries(received.toRecord())) {
    receivedObj[path] = normalizeMempostEntry(path, buf);
  }
  if (expected instanceof Mempost) {
    for (const [path, buf] of Object.entries(expected.toRecord())) {
      expectedObj[path] = normalizeMempostEntry(path, buf);
    }
  } else {
    for (const [path, contents] of Object.entries(expected)) {
      expectedObj[path] = normalizeMempostEntry(path, contents);
    }
  }

  if (this.isNot) {
    expect(receivedObj).not.toEqual(expectedObj);
  } else {
    expect(receivedObj).toEqual(expectedObj);
  }
  return {
    pass: !this.isNot,
    message: () => 'Matched',
  };
}

function toEqualHTML(
  this: jest.MatcherContext,
  received: unknown,
  expected: string
): jest.CustomMatcherResult {
  if (!isString(received)) {
    return {
      pass: false,
      message: () => `Expected to receive string type but had ${received}`,
    };
  }
  const normalRecv = normalizeHTML(received);
  const normalExpected = normalizeHTML(expected);
  if (this.isNot) {
    expect(normalRecv).not.toEqual(normalExpected);
  } else {
    expect(normalRecv).toEqual(normalExpected);
  }
  return {
    pass: !this.isNot,
    message: () => 'Matched',
  };
}

export const ALL_JEST_MATCHERS: jest.ExpectExtendMap = {
  toEqualHTML,
  toEqualMempost,
};
