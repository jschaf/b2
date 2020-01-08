import { Doc } from '//post/hast/doc';
import * as h from '//post/hast/nodes';
import * as hast from 'hast-format';

describe('Doc', () => {
  const testData: [string, Doc, h.RootContent[], hast.Root][] = [
    [
      'empty',
      Doc.create(),
      [],
      h.root([
        h.doctype(),
        h.elemProps('html', { lang: 'en' }, [
          h.elem('head', []),
          h.elem('body', []),
        ]),
      ]),
    ],
    [
      'with children',
      Doc.create(),
      [h.elemText('p', 'alpha')],
      h.root([
        h.doctype(),
        h.elemProps('html', { lang: 'en' }, [
          h.elem('head', []),
          h.elem('body', [h.elemText('p', 'alpha')]),
        ]),
      ]),
    ],
    [
      'with title',
      Doc.create().addToHead(h.elemText('title', 'alpha')),
      [],
      h.root([
        h.doctype(),
        h.elemProps('html', { lang: 'en' }, [
          h.elem('head', [h.elemText('title', 'alpha')]),
          h.elem('body', []),
        ]),
      ]),
    ],
  ];

  for (const [name, doc, children, expected] of testData) {
    it(name, () => {
      const d = doc.toHast(children);
      expect(d).toEqual(expected);
    });
  }
});
