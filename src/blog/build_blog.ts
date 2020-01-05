import * as files from '//files';
import { PostCompiler } from '//post/compiler';
import { PostAST } from '//post/ast';
import * as path from 'path';
import * as fs from 'fs';
import { PostBag } from '//post/post_bag';

const buildBlog = async (): Promise<void> => {
  const gitDir = files.findGitDirectory(__dirname);
  const rootDir = path.dirname(gitDir);
  const postsDir = path.join(rootDir, 'posts');
  // Find bare files
  const postCompiler = PostCompiler.create();

  const markdowns = await fs.promises.readdir(postsDir);
  await Promise.all(
    markdowns.map(
      async (mdPath): Promise<void> => {
        if (
          path.extname(mdPath) !== '.md' ||
          path.basename(mdPath) === 'index.md'
        ) {
          console.log('!!! Skipping because not .md file or is index.md');
          return;
        }
        const buf = await fs.promises.readFile(path.join(postsDir, mdPath));
        const md = buf.toString('utf8');
        const postBag = PostBag.fromTomlFrontmatterMarkdown(md);
        const postAST = PostAST.fromMdast(postBag.postNode.node);
        const mp = postCompiler.compileToMempost(postAST);
        const slug = (postBag.postNode.metadata.schema[
          'slug'
        ] as unknown) as string;
        const outDir = path.join(rootDir, 'public', slug, 'index.html');
        console.log('!!! outDir', outDir);
        await fs.promises.mkdir(path.dirname(outDir), { recursive: true });
        await fs.promises.writeFile(outDir, mp.getEntry('index.html'));
      }
    )
  );

  // Find dir files
  console.log('!!! postsDir', postsDir);
};

if (require.main === module) {
  buildBlog();
}
