import * as memfs from 'memfs';
import { dedent } from '../strings';
import { PostBag } from './post_bag';
import { PostCommitter } from './post_committer';

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

    await PostCommitter.forFs(memfs.createFsFromVolume(vol)).commit(
      '/root',
      bag
    );

    expect(vol.toJSON()).toEqual({
      '/root/posts/foo_bar.md': '# Hello\n',
    });
  });
});
