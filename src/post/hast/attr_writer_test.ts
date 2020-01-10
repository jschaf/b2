import { AttrWriter } from '//post/hast/attr_writer';
import { StringBuilder } from '//strings';

describe('AttrWriter', () => {
  const testData: [string, Record<string, any>, string][] = [
    ['1 string', { foo: 'bar' }, 'foo="bar"'],
    ['1 number', { foo: 1 }, 'foo="1"'],
    ['string - &quot;', { foo: '"qux' }, 'foo="&quot;qux"'],
    ['array - &gt;', { foo: [1, 'foo', '>baz<'] }, 'foo="1 foo &gt;baz&lt;"'],
    ['empty string', { foo: '' }, 'foo'],
    ['null value', { foo: null }, 'foo'],
    ['true value', { foo: true }, 'foo'],
    ['false value', { foo: false }, ''],
    ['date value', { foo: new Date('2019-01-01') }, 'foo'],
    ['bad name - &', { 'foo&': '' }, ''],
    ['save ampersand in value', { foo: '&bar' }, 'foo="&bar"'],
    ['array', { foo: [1, 'alpha', 'bravo'] }, 'foo="1 alpha bravo"'],
    ['2 props', { foo: '&bar', baz: 'qux' }, 'foo="&bar" baz="qux"'],
  ];
  for (const [name, props, expected] of testData) {
    it(name, () => {
      const sb = StringBuilder.create();
      AttrWriter.create().writeElemProps(props, sb);
      expect(sb.toString()).toEqual(expected);
    });
  }
});
