/**
 * This package contains custom Jest matchers.
 *
 * See https://stackoverflow.com/a/45745719/30900 for an example of how to write
 * a custom matcher.
 *
 * Update global_jest_matchers.d.ts for Typescript type-checking.
 */

import * as util from 'util';
import { Mempost } from '../post/mempost';
import unified from 'unified';
import rehypeParse from 'rehype-parse';
import rehypeFormat from 'rehype-format';
import rehypeStringify from 'rehype-stringify';

/**
 * Converts a Buffer to a UTF-8 string if possible. Otherwise, return the buffer.
 *
 * Intended purposed is to produce cleaner error messages.
 */
const normalizeMempostBuffer = (path: string, buf: Buffer): string | Buffer => {
  const td = new util.TextDecoder('utf8', { fatal: true });
  try {
    const str = td.decode(buf);
    return normalizeMempostString(path, str);
  } catch (e) {
    return buf;
  }
};

const normalizeMempostString = (path: string, contents: string): string => {
  if (path.endsWith('.html')) {
    const vFile = unified()
      .use(rehypeParse)
      .use(rehypeFormat)
      .use(rehypeStringify)
      .processSync(contents);
    return vFile.contents.toString('utf8');
  }
  return contents;
};

function toEqualMempost(
  this: jest.MatcherContext,
  received: any,
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
  for (const [path, buf] of received.entries()) {
    receivedObj[path] = normalizeMempostBuffer(path, buf);
  }
  if (expected instanceof Mempost) {
    for (const [path, buf] of expected.entries()) {
      expectedObj[path] = normalizeMempostBuffer(path, buf);
    }
  } else {
    for (const [path, contents] of Object.entries(expected)) {
      expectedObj[path] = normalizeMempostString(path, contents);
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

export const ALL_JEST_MATCHERS: jest.ExpectExtendMap = { toEqualMempost };
