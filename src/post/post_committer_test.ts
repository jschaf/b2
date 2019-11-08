import {PostBag} from '//post/post_bag';
import {PostCommitter} from '//post/post_committer';
import {withDefaultFrontMatter} from 'front_matters.ts';
import {dedent} from '//strings';
import * as memfs from 'memfs';
import * as path from 'path';

describe('PostCommitter', () => {
  it('should commit a standalone post', async () => {
    const bag = PostBag.fromMarkdown(withDefaultFrontMatter(dedent`
      # Hello
    `));
    const vol = new memfs.Volume();
    const fileSystem = memfs.createFsFromVolume(vol);
    const gitDir = '/root';

    await PostCommitter.forFs(fileSystem, gitDir).commit(bag);

    expect(removeGit(gitDir, vol.toJSON())).toEqual({
      '/root/posts/foo_bar.md': '# Hello\n',
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
