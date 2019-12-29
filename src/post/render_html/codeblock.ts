import {mdCode} from '//post/testing/markdown_nodes';
import {isString} from '//strings';
import * as unist from 'unist';
import vfile from 'vfile';

//import * as prism from 'prismjs';

interface B2Transformer {
  transformSync(n: unist.Node, vf: vfile.VFile): Error | unist.Node | void
}

export class CodeblockRenderer implements B2Transformer {
  private constructor() {
  }

  static create(): CodeblockRenderer {
    return new CodeblockRenderer();
  }

  transformSync(n: unist.Node, _vf: vfile.VFile): Error | unist.Node | void {
    if (isCodeblockNode(n)) {
      return mdCode('joe was here')
    }
  }
}

const isCodeblockNode =
    (n: unist.Node): n is { type: 'code'; value: string } => {
      return n.type === 'code' && isString(n.value);
    };
