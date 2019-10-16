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

  parse(markdown: string): unist.Node {
    return this.processor.parse(markdown);
  }
}
