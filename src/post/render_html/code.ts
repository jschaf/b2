import { htmlNode } from '//post/render_html/hast_nodes';
import { HastRenderer } from '//post/render_html/render';
import { isString } from '//strings';
import * as unist from 'unist';
import * as mdast from 'mdast';
import vfile from 'vfile';

export class CodeRenderer implements HastRenderer {
  private constructor() {}

  static create(): CodeRenderer {
    return new CodeRenderer();
  }

  render(node: mdast.Code, _vf: vfile.VFile): Error | unist.Node {
    const props: Record<string, unknown> = {};
    if (isString(node.lang) && node.lang !== '') {
      props.className = ['lang-' + node.lang.trim()];
    }
    return htmlNode('pre', {}, [
      htmlNode('code', props, [{ type: 'text', value: node.value }]),
    ]);
  }
}
