// Parser for post content.
import unified from 'unified';
import * as unist from 'unist';
import remarkParse from 'remark-parse';
import nodeRemove from 'unist-util-remove';
import { PostMetadata } from './post_metadata';

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
    const metadata = PostMetadata.parseFromMarkdownAST(node);
    const nodeSansMetadata = nodeRemove(node, PostMetadata.isMetadataNode) || {
      type: 'root',
    };
    return new PostNode(metadata, nodeSansMetadata);
  }
}

export class PostNode {
  constructor(readonly metadata: PostMetadata, readonly node: unist.Node) {}
}
