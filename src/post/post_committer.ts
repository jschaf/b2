import * as fs from 'fs';
import * as memfs from 'memfs';
import { GitFsPlugin } from 'isomorphic-git';
import * as isoGit from 'isomorphic-git';
import * as path from 'path';
import remarkStringify from 'remark-stringify';
import unified from 'unified';
import { checkArg, checkDefinedAndNotNull } from '../asserts';
import { Mempost } from './mempost';
import { PostBag } from './post_bag';

export class PostCommitter {
  private constructor(
    private readonly fs: FsModule,
    private readonly dir: string
  ) {
    isoGit.plugins.set('fs', this.fs);
  }

  static forFs(fs: FsModule, dir: string): PostCommitter {
    checkArg(
      path.isAbsolute(dir),
      `Expected Git dir to be an absolute path but had '${dir}'.`
    );
    return new PostCommitter(fs, dir);
  }

  /**
   * Commits the source files of the post bag onto the filesystem
   * relative to dir.
   */
  async commit(bag: PostBag): Promise<void> {
    await isoGit.init({ dir: this.dir, noOverwrite: true });

    const mempost = await PostSrcRenderer.create().render(bag);
    for (const [relPath, contents] of mempost.entries()) {
      const fullPath = path.resolve(this.dir, relPath);
      await this.fs.promises.mkdir(path.dirname(fullPath), { recursive: true });
      await this.fs.promises.writeFile(fullPath, contents);
      await isoGit.add({ dir: this.dir, filepath: relPath });
    }
    await isoGit.commit({
      dir: this.dir,
      message: `auto: Edit ${bag.postNode.metadata.slug}`,
      author: {
        name: 'Joe Schafer',
        email: 'joe@schafer.dev',
        date: new Date(),
      },
    });
  }

  async pushOrigin(): Promise<void> {
    const token = checkDefinedAndNotNull(
      process.env['GITHUB_TOKEN'],
      'No environment variable GITHUB_TOKEN'
    );
    await isoGit.push({
      dir: this.dir,
      remote: 'origin',
      ref: 'master',
      username: 'jschaf',
      token,
    });
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

type FsModule = GitFsPlugin & (typeof fs | memfs.IFs);
