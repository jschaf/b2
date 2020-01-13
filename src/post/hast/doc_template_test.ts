import { DocTemplate } from '//post/hast/doc_template';
import * as h from '//post/hast/nodes';

describe('Doc', () => {
  const testData: [string, DocTemplate, h.RootContent[]][] = [
    ['empty', DocTemplate.create(), []],
    [
      'with children',
      DocTemplate.create(),
      [h.elem('body', [h.elemText('p', 'alpha')])],
    ],
    [
      'with title',
      DocTemplate.create().addToHead(h.elemText('title', 'alpha')),
      [],
    ],
  ];

  for (const [name, doc, children] of testData) {
    it(name, () => {
      const d = doc.render(children);
      expect(d).toMatchSnapshot();
    });
  }
});
