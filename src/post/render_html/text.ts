import { hastText } from '//post/render_html/hast_nodes';
import { checkNodeType } from '//post/render_html/md_nodes';
import { HastRenderer } from '//post/render_html/render';
import * as unist from 'unist';
import * as md_nodes from '//post/render_html/md_nodes';
import vfile from 'vfile';

export class TextRenderer implements HastRenderer {
  private constructor() {}

  static create(): TextRenderer {
    return new TextRenderer();
  }

  render(node: unist.Node, _vf: vfile.VFile): Error | unist.Node {
    checkNodeType(node, 'text', md_nodes.isText);
    return hastText(node.value);
  }
}
