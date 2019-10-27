// Parser for post content.
import unified from 'unified';
import * as unist from 'unist';
import remarkParse from "remark-parse";
import {PostMetadata} from "./post_metadata";
import * as dates from '../dates';

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
    return new PostNode(PostMetadata.of({
      slug: 'fixme',
      date: dates.fromISO('2019-10-20'),
    }), node);
  }
}

export class PostNode {
  constructor(
      readonly metadata: PostMetadata,
      readonly node: unist.Node) {
  }
}
