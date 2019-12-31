import { RenderDispatcher } from '//post/render_html/dispatch';
import { htmlNode } from '//post/render_html/hast_nodes';
import { checkNodeType } from '//post/render_html/md_nodes';
import { HastRenderer } from '//post/render_html/render';
import { renderChildren } from '//post/render_html/renders';
import * as unist from 'unist';
import * as md_nodes from '//post/render_html/md_nodes';
import vfile from 'vfile';

export class HeadingRenderer implements HastRenderer {
  private constructor(private readonly dispatcher: RenderDispatcher) {}

  static create(dispatcher: RenderDispatcher): HeadingRenderer {
    return new HeadingRenderer(dispatcher);
  }

  render(node: unist.Node, vf: vfile.VFile): Error | unist.Node {
    checkNodeType(node, 'heading', md_nodes.isHeading);
    console.log('!!! Rendering heading: ' + node.depth);
    const childRenders = renderChildren(node, vf, this.dispatcher);
    return htmlNode('h' + node.depth, {}, childRenders);
  }
}
