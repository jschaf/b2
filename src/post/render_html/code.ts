import {
  hastElem,
  hastElemWithProps,
  hastText,
} from '//post/render_html/hast_nodes';
import { checkNodeType } from '//post/render_html/md_nodes';
import * as md_nodes from '//post/render_html/md_nodes';
import { HastRenderer } from '//post/render_html/render';
import { isString } from '//strings';
import * as unist from 'unist';
import vfile from 'vfile';

export class CodeRenderer implements HastRenderer {
  private constructor() {}

  static create(): CodeRenderer {
    return new CodeRenderer();
  }

  render(node: unist.Node, _vf: vfile.VFile): Error | unist.Node {
    checkNodeType(node, 'code', md_nodes.isCode);
    const props: Record<string, unknown> = {};
    if (isString(node.lang) && node.lang !== '') {
      props.className = ['lang-' + node.lang.trim()];
    }
    return hastElem('pre', [
      hastElemWithProps('code', props, [hastText(node.value)]),
    ]);
  }
}
