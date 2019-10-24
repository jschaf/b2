// Parser for post content.
import unified from 'unified';
import * as unist from 'unist';
import remarkParse from "remark-parse";

export class PostParser {
  private readonly processor: unified.Processor<unified.Settings>;

  private constructor() {
    this.processor = unified().use(remarkParse);
  }

  static create(): PostParser {
    return new PostParser();
  }

  parse(markdown: string): PostNode {
    const node = this.processor.parse(markdown);
    return new PostNode({}, node);
  }
}

export class PostNode {
  constructor(
      readonly frontMatter: unknown,
      readonly node: unist.Node) {
  }
}
