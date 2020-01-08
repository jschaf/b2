import * as hast from 'hast-format';
import * as unist from 'unist';
import * as h from '//post/hast/nodes';

/** Doc is an hast template for a full HTML document. */
export class Doc {
  private readonly head: unist.Node[] = [];
  private constructor() {}

  static create(): Doc {
    return new Doc();
  }

  addToHead(...hs: unist.Node[]): Doc {
    this.head.push(...hs);
    return this;
  }

  toHast(children: h.RootContent[]): hast.Root {
    return h.root([
      h.doctype(),
      h.elemProps('html', { lang: 'en' }, [
        h.elem('head', this.head),
        h.elem('body', children),
      ]),
    ]);
  }
}
