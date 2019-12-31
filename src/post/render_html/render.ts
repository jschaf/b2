import { checkDefined } from '//asserts';
import { Mempost } from '//post/mempost';
import { PostBag } from '//post/post_bag';
import { RenderDispatcher } from '//post/render_html/dispatch';
//import rehypeDocument from 'rehype-document';
//import rehypeKatex from 'rehype-katex';
import rehypeStringify from 'rehype-stringify';
//import remarkRehype from 'remark-rehype';
import unified from 'unified';
import * as unist from 'unist';
import vfile from 'vfile';

export interface HastRenderer {
  render(tree: unist.Node, vf: vfile.VFile): Error | unist.Node;
}

/** Compiles a markdown AST (mdast) into an HTML AST (hast). */
class HastCompiler {
  private constructor(private readonly dispatcher: RenderDispatcher) {}

  static create(): HastCompiler {
    return new HastCompiler(RenderDispatcher.defaultInstance());
  }

  compile(mdast: unist.Node, vf: vfile.VFile): Error | unist.Node | void {
    const renderer = checkDefined(
      this.dispatcher.dispatch(mdast),
      `No renderer exists for node type ${mdast.type}`
    );
    return renderer.render(mdast, vf);
  }

  asUnifiedAttacher(): unified.Attacher<[], {}> {
    const compilerThis = this;
    return function(this: unified.Processor<{}>): unified.Transformer {
      return compilerThis.compile.bind(compilerThis);
    };
  }
}

/**
 * Renders a post and all required assets onto a new Mempost.
 */
export class PostHtmlRenderer {
  private readonly processor: unified.Processor<unified.Settings>;

  private constructor() {
    this.processor = unified()
      .use(HastCompiler.create().asUnifiedAttacher())
      .use(rehypeStringify);
  }

  static create(): PostHtmlRenderer {
    return new PostHtmlRenderer();
  }

  /**
   * Renders a post from an existing Mempost into an HTML representation on a new Mempost.
   */
  async render(bag: PostBag): Promise<Mempost> {
    const dest = Mempost.create();
    const htmlNode = await this.processor.run(bag.postNode.node);
    const htmlString = this.processor.stringify(htmlNode);
    dest.addUtf8Entry('index.html', htmlString);
    return dest;
  }
}
