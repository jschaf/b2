import { HTMLEscaper } from '//post/html/escapers';

describe('HTMLEscaper', () => {
  describe('escape', () => {
    const testData: [string, string, string][] = [
      ['repeated ampersand', '&a;foo&b;', '&amp;a;foo&amp;b;'],
      ['okay ampersand', '&foo&', '&foo&'],
      ['quotes', '"foo"', '&quot;foo&quot;'],
      ['single quotes', "'foo'", '&#39;foo&#39;'],
      ['less than', '<foo<', '&lt;foo&lt;'],
      ['greater than', '>foo>', '&gt;foo&gt;'],
    ];
    for (const [name, input, expected] of testData) {
      it(name, () => {
        const h = HTMLEscaper.create().escape(input);
        expect(h).toEqual(expected);
      });
    }
  });

  describe('needsEscape', () => {
    const testData: [string, string, boolean][] = [
      ['safe ampersand', 'foo&', false],
      ['ambiguous ampersand', 'foo&a;', true],
      ['quotes', "foo'", true],
      ['quotes', 'foo"', true],
      ['greater than', 'fo>o', true],
      ['less than', 'foo<', true],
      ['none', 'foo', false],
      ['none', 'false', false],
    ];
    for (const [name, input, expected] of testData) {
      it(name, () => {
        const h = HTMLEscaper.create().needsEscaped(input);
        expect(h).toEqual(expected);
      });
    }
  });
});
