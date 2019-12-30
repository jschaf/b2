import rehypeStringify from 'rehype-stringify';
import rehypeKatex from 'rehype-katex';
import remarkRehype from 'remark-rehype';
import rehypeDocument from 'rehype-document';
import unified from 'unified';
import { Mempost } from '//post/mempost';
import { PostBag } from '//post/post_bag';

/**
 * Renders a post and all required assets onto a new Mempost.
 */
export class PostHtmlRenderer {
  private readonly processor: unified.Processor<unified.Settings>;

  private constructor() {
    this.processor = unified()
      .use(remarkRehype)
      .use(rehypeKatex)
      .use(rehypeDocument)
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
