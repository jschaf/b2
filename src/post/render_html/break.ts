import { hastElem } from '//post/render_html/hast_nodes';
import { checkNodeType } from '//post/render_html/md_nodes';
import { HastRenderer } from '//post/render_html/render';
import * as unist from 'unist';
import * as md_nodes from '//post/render_html/md_nodes';
import vfile from 'vfile';

export class BreakRenderer implements HastRenderer {
  private constructor() {}

  static create(): BreakRenderer {
    return new BreakRenderer();
  }

  render(node: unist.Node, _vf: vfile.VFile): Error | unist.Node {
    checkNodeType(node, 'break', md_nodes.isBreak);
    return hastElem('break');
  }
}
