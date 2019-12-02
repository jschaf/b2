import { PostMetadata } from '//post/post_metadata';
import * as dates from '//dates';
import {
  mdCode,
  mdFrontmatterToml,
  mdHeading1,
  mdRoot,
} from '//post/testing/markdown_nodes';
import { dedent } from '//strings';

test('parses valid tokens', () => {
  const slug = 'qux_bar';
  let date = '2017-06-18';
  const tree = mdRoot([
    mdHeading1('hello'),
    mdCode(dedent`
        # Metadata
        slug: ${slug}
        date: ${date}
      `),
  ]);

  const metadata = PostMetadata.parseFromMarkdownAST(tree);

  expect(metadata.schema).toEqual({ slug, date: dates.fromISO(date) });
});

test('parses valid tokens from toml frontmatter', () => {
  const slug = 'qux_bar';
  let date = dates.fromISO('2019-05-18');
  const tree = mdRoot([mdFrontmatterToml({ slug, date })]);

  const metadata = PostMetadata.parseFromTomlFrontmatter(tree);

  expect(metadata.schema).toEqual({ slug, date });
});
