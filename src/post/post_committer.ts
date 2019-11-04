import * as memfs from 'memfs';
import * as path from 'path';
import remarkStringify from 'remark-stringify';
import unified from 'unified';
import { Mempost } from './mempost';
import { PostBag } from './post_bag';

export class PostCommitter {
  private constructor(private readonly fs: memfs.IFs) {}

  static forFs(fs: memfs.IFs): PostCommitter {
    return new PostCommitter(fs);
  }

  /**
   * Commits the source files of the post bag onto the filesystem
   * relative to dir.
   */
  async commit(dir: string, bag: PostBag): Promise<void> {
    const mempost = await PostSrcRenderer.create().render(bag);
    for (const [relPath, contents] of mempost.entries()) {
      const fullPath = path.resolve(dir, relPath);
      await this.fs.promises.mkdir(path.dirname(fullPath), { recursive: true });
      await this.fs.promises.writeFile(fullPath, contents);
    }
  }
}

/**
 * Renders the source view of a post. This is the view used to commit
 * to the repository.
 */
class PostSrcRenderer {
  private readonly processor: unified.Processor<unified.Settings>;

  private constructor() {
    this.processor = unified().use(remarkStringify);
  }

  static create() {
    return new PostSrcRenderer();
  }

  async render(bag: PostBag): Promise<Mempost> {
    const md = this.processor.stringify(bag.postNode.node);
    return Mempost.ofUtf8Entry(`posts/${bag.postNode.metadata.slug}.md`, md);
  }
}
