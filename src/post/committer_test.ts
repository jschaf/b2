import { PostAST } from '//post/ast';
import * as md from '//post/mdast/nodes';
import { PostCommitter } from '//post/committer';
import * as frontMatters from '//post/testing/front_matters';
import { dedent } from '//strings';
import * as memfs from 'memfs';
import * as path from 'path';

describe('PostCommitter', () => {
  it('should commit a standalone post', async () => {
    const ast = PostAST.fromMdast(
      md.root([
        frontMatters.defaultTomlMdast(),
        md.headingText('h1', 'alpha'),
        md.paragraphText('Foo bar.'),
      ])
    );

    const vol = new memfs.Volume();
    const fileSystem = memfs.createFsFromVolume(vol);
    const gitDir = '/root';

    await PostCommitter.forFs(fileSystem, gitDir).commit(ast);

    expect(removeGit(gitDir, vol.toJSON())).toEqual({
      '/root/posts/foo_bar.md': trailingNewline(dedent`
          ${frontMatters.defaultTomlBlock()}

          # alpha

          Foo bar.
      `),
    });
  });
});

const removeGit = (
  dir: string,
  files: Record<string, string | null>
): Record<string, string | null> => {
  const nonGitFiles: Record<string, string | null> = {};
  const gitDir = path.resolve(dir, '.git');
  for (const [filePath, content] of Object.entries(files)) {
    if (!filePath.startsWith(gitDir)) {
      nonGitFiles[filePath] = content;
    }
  }
  return nonGitFiles;
};

const trailingNewline = (s: string): string => {
  return s + '\n';
};
