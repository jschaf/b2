import { PostAST } from '//post/ast';
import { MarkdownParser } from '//post/markdown/parser';
import { checkState } from '//asserts';
import { Unzipper } from '//zip_files';

export const TEXT_PACK_BUNDLE_PREFIX = 'Content.textbundle';

/** Parser for post content. */
export class PostParser {
  private readonly mdParser: MarkdownParser;

  private constructor() {
    this.mdParser = MarkdownParser.create();
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
    return this.mdParser.parse(markdown);
  }
}
