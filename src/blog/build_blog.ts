import * as files from '//files';
import { PostCompiler } from '//post/compiler';
import { PostParser } from '//post/parser';
import * as fs from 'fs';
import * as path from 'path';

const buildBlog = async (): Promise<void> => {
  const gitDir = files.findGitDirectory(__dirname);
  const rootDir = path.dirname(gitDir);
  const postsDir = path.join(rootDir, 'posts');
  // Find bare files
  const postParser = PostParser.create();
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
        const ast = postParser.parseMarkdown(md);
        const cp = postCompiler.compile(ast);
        const outDir = path.join(rootDir, 'public');
        await cp.write(outDir);
      }
    )
  );

  // Find dir files
  console.log('!!! postsDir', postsDir);
};

if (require.main === module) {
  buildBlog();
}
