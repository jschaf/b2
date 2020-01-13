import { DocTemplate } from '//post/hast/doc_template';
import * as h from '//post/hast/nodes';
import * as hast from 'hast-format';

describe('Doc', () => {
  const testData: [string, DocTemplate, h.RootContent[], hast.Root][] = [
    [
      'empty',
      DocTemplate.create(),
      [],
      h.root([
        h.doctype(),
        h.elemProps('html', { lang: 'en' }, [h.elem('head', [])]),
      ]),
    ],
    [
      'with children',
      DocTemplate.create(),
      [h.elem('body', [h.elemText('p', 'alpha')])],
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
      DocTemplate.create().addToHead(h.elemText('title', 'alpha')),
      [],
      h.root([
        h.doctype(),
        h.elemProps('html', { lang: 'en' }, [
          h.elem('head', [h.elemText('title', 'alpha')]),
        ]),
      ]),
    ],
  ];

  for (const [name, doc, children, expected] of testData) {
    it(name, () => {
      const d = doc.render(children);
      expect(d).toEqual(expected);
    });
  }
});
