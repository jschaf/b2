// Utility methods for working with renderers.

import * as unist from 'unist';

import { RenderDispatcher } from '//post/render_html/dispatch';
import vfile from 'vfile';

/** Renders node.children using the renderer determined by dispatcher. */
export const renderChildren = (
  parent: unist.Parent,
  vf: vfile.VFile,
  rd: RenderDispatcher
): unist.Node[] => {
  const results: unist.Node[] = [];
  for (const child of parent.children) {
    const r = rd.dispatch(child);
    if (r === undefined) {
      console.log(`Unable to find renderer for node type: ${child.type}`);
      continue;
    }
    const result = r.render(child, vf);
    if (result instanceof Error) {
      console.log(`Got error: ${result}`);
      continue;
    }

    results.push(result);
  }
  return results;
};
