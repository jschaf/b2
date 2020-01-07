import { PostAST } from '//post/ast';
import remarkFrontmatter from 'remark-frontmatter';
import remarkParse from 'remark-parse';
import unified from 'unified';

/** Parses a markdown document into mdast. */
export class MarkdownParser {
  private readonly processor: unified.Processor;

  private constructor() {
    this.processor = unified()
      .use(remarkParse, { commonmark: true })
      .use(remarkFrontmatter, ['toml']);
  }

  static create(): MarkdownParser {
    return new MarkdownParser();
  }

  parse(markdown: string): PostAST {
    const node = this.processor.parse(markdown);
    return PostAST.fromMdast(node);
  }
}
