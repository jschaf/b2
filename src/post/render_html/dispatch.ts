import { CodeRenderer } from '//post/render_html/code';
import { HeadingRenderer } from '//post/render_html/heading';
import { ParagraphRenderer } from '//post/render_html/paragraph';
import { HastRenderer } from '//post/render_html/render';
import { RootRenderer } from '//post/render_html/root';
import { TextRenderer } from '//post/render_html/text';
import * as unist from 'unist';

/** Chooses the correct renderer for a node. */
export class RenderDispatcher {
  private static instance: RenderDispatcher | null = null;

  private readonly renderersByType: Map<string, HastRenderer>;

  private constructor(renderers: Map<string, HastRenderer>) {
    this.renderersByType = renderers;
  }

  static defaultInstance(): RenderDispatcher {
    if (RenderDispatcher.instance === null) {
      RenderDispatcher.instance = new RenderDispatcher(new Map());
      const rd = RenderDispatcher.instance;
      const defaults: [string, HastRenderer][] = [
        ['code', CodeRenderer.create()],
        ['heading', HeadingRenderer.create(rd)],
        ['paragraph', ParagraphRenderer.create(rd)],
        ['text', TextRenderer.create()],
        ['root', RootRenderer.create(rd)],
      ];
      for (const [type, r] of defaults) {
        rd.renderersByType.set(type, r);
      }
    }
    return RenderDispatcher.instance;
  }

  dispatch(node: unist.Node): HastRenderer | undefined {
    const r = this.renderersByType.get(node.type);
    const name = r === undefined ? '<undefined>' : r.constructor.name;
    console.log(`!!! Dispatching node: ${node.type} to ${name}`);
    return r;
  }
}
