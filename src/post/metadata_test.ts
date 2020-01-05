import * as md from '//post/mdast/nodes';
import { PostMetadata } from '//post/metadata';
import * as dates from '//dates';
import { dedent } from '//strings';

test('parses valid tokens', () => {
  const slug = 'qux_bar';
  let date = '2017-06-18';
  const tree = md.root([
    md.headingText('h1', 'hello'),
    md.code(dedent`
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
  const tree = md.root([md.tomlFrontmatter({ slug, date })]);

  const metadata = PostMetadata.parseFromTomlFrontmatter(tree);

  expect(metadata.schema).toEqual({ slug, date });
});
