import * as memfs from 'memfs';
import * as path from 'path';
import { dedent } from '//strings';
import { PostBag } from '//post/post_bag';
import { PostCommitter } from '//post/post_committer';

describe('PostCommitter', () => {
  it('should commit a standalone post', async () => {
    const bag = PostBag.fromMarkdown(dedent`
      # Hello
      
      \`\`\`yaml
      # Metadata
      slug: foo_bar
      date: 2019-10-08
      \`\`\`
    `);
    const vol = new memfs.Volume();

    const gitDir = '/root';
    await PostCommitter.forFs(memfs.createFsFromVolume(vol), gitDir).commit(
      bag
    );

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
