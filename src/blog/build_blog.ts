import * as files from '//files';
import { PostRenderer } from '//post/post_renderer';
import * as path from 'path';
import * as fs from 'fs';
import { PostBag } from '//post/post_bag';
import { Mempost } from '//post/mempost';

const buildBlog = async (): Promise<void> => {
  const gitDir = files.findGitDirectory(__dirname);
  const rootDir = path.dirname(gitDir);
  const postsDir = path.join(rootDir, 'posts');
  // Find bare files
  const postRenderer = PostRenderer.create();
  const markdowns = await fs.promises.readdir(postsDir);
  const memposts = await Promise.all(
    markdowns
      .map((mdPath): null | Promise<Mempost> => {
        if (
          path.extname(mdPath) !== '.md' ||
          path.basename(mdPath) === 'index.md'
        ) {
          return null;
        }
        const md = fs
          .readFileSync(path.join(postsDir, mdPath))
          .toString('utf8');
        const postBag = PostBag.fromTomlFrontmatterMarkdown(md);
        return postRenderer.render(postBag);
      })
      .filter(x => x !== null)
  );

  // Find dir files
  console.log('!!! postsDir', postsDir);
  console.log('!!! memposts', memposts);
};

if (require.main === module) {
  buildBlog();
}
