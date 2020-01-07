import { PostAST } from '//post/ast';
import remarkParse from 'remark-parse';
import remarkFrontmatter from 'remark-frontmatter';
import unified from 'unified';
import { checkState } from '//asserts';
import { Unzipper } from '//zip_files';

export const TEXT_PACK_BUNDLE_PREFIX = 'Content.textbundle';

/** Parser for post content. */
export class PostParser {
  private readonly processor: unified.Processor;

  private constructor() {
    this.processor = unified()
      .use(remarkParse, { commonmark: true })
      .use(remarkFrontmatter, ['toml']);
  }

  static create(): PostParser {
    return new PostParser();
  }

  async parseTextPack(textPack: Buffer): Promise<PostAST> {
    const entries = await Unzipper.unzip(textPack);
    const texts = entries.filter(
      e => e.filePath === TEXT_PACK_BUNDLE_PREFIX + '/text.md'
    );
    checkState(
      texts.length === 1,
      'Expected a single text.md file in TextPack.'
    );
    const text = texts[0];
    return this.parseMarkdown(text.contents.toString('utf8'));
  }

  parseMarkdown(markdown: string): PostAST {
    const node = this.processor.parse(markdown);
    return PostAST.fromMdast(node);
  }
}
