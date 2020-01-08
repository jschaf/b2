import * as h from '//post/hast/nodes';
import { PostType } from '//post/metadata';
import * as hast from 'hast-format';
import * as unist from 'unist';

/** DocTemplate is an hast template for a full HTML document. */
export class DocTemplate {
  private readonly head: unist.Node[] = [];
  private constructor() {}

  static create(): DocTemplate {
    return new DocTemplate();
  }

  static templates(): Map<PostType, DocTemplate> {
    return new Map<PostType, DocTemplate>(templates);
  }

  addToHead(...hs: unist.Node[]): DocTemplate {
    this.head.push(...hs);
    return this;
  }

  render(children: unist.Node[]): hast.Root {
    return h.root([
      h.doctype(),
      h.elemProps('html', { lang: 'en' }, [
        h.elem('head', this.head),
        h.elem('body', children),
      ]),
    ]);
  }
}

const templates: [PostType, DocTemplate][] = [
  [PostType.Post, DocTemplate.create()],
];
