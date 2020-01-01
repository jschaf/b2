import { RenderDispatcher } from '//post/render_html/dispatch';
import { hastElem } from '//post/render_html/hast_nodes';
import { HastRenderer } from '//post/render_html/render';
import { renderChildren } from '//post/render_html/renders';
import * as unist from 'unist';
import * as mdast from 'mdast';
import vfile from 'vfile';

export class RootRenderer implements HastRenderer {
  private constructor(private readonly dispatcher: RenderDispatcher) {}

  static create(dispatch: RenderDispatcher): RootRenderer {
    return new RootRenderer(dispatch);
  }

  render(node: mdast.Root, vf: vfile.VFile): Error | unist.Node {
    const results = renderChildren(node, vf, this.dispatcher);
    return hastElem('body', results);
  }
}
