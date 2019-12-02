import {PostMetadata} from '//post/post_metadata';
import {
  PostNode,
  PostParser,
  TEXT_PACK_BUNDLE_PREFIX,
} from '//post/post_parser';
import {
  DEFAULT_FRONTMATTER,
  withDefaultFrontMatter,
} from '//post/testing/front_matters';
import {dedent} from '//strings';
import {ZipFileEntry, Zipper} from '//zip_files';
import {
  mdHeading,
  mdOrderedList,
  mdParaText,
  mdRoot,
  mdText, mdFrontmatterToml,
  stripPositions, mdHeading1,
} from '//post/testing/markdown_nodes';
import * as dates from '//dates';

test('parses from markdown', () => {
  const markdown = withDefaultFrontMatter(dedent`
    # hello
    
    Hello world.
   `);
  const node = PostParser.create().parseMarkdown(markdown);

  const expected = mdRoot([
    mdHeading(1, [mdText('hello')]),
    mdParaText('Hello world.'),
  ]);
  expect(stripPositions(node)).toEqual(
      new PostNode(DEFAULT_FRONTMATTER, expected)
  );
});

test('parses paragraph followed immediately by a list', () => {
  const markdown = withDefaultFrontMatter(dedent`
    Hello world.
    1. text
  `);
  const node = PostParser.create().parseMarkdown(markdown);

  const expected = mdRoot([
    mdParaText('Hello world.'),
    mdOrderedList([mdParaText('text')]),
  ]);
  expect(stripPositions(node).node).toEqual(expected);
});

test('parses from TextPack', async () => {
  const markdown = withDefaultFrontMatter(dedent`
    # hello
    
    Hello world.
  `);
  const buf = await Zipper.zip([
    ZipFileEntry.ofUtf8(TEXT_PACK_BUNDLE_PREFIX + '/text.md', markdown),
  ]);
  const node = await PostParser.create().parseTextPack(buf);

  const expected = mdRoot([
    mdHeading(1, [mdText('hello')]),
    mdParaText('Hello world.'),
  ]);
  expect(stripPositions(node)).toEqual(
      new PostNode(DEFAULT_FRONTMATTER, expected)
  );
});

test('parses from frontmatter markdown', async () => {
  const slug = "foo_qux";
  const date = '2019-10-17';
  const markdown = dedent`
    +++
    slug = "${slug}"
    date = ${date}
    +++
    
    # Hello
  `;

  const node = PostParser.create().parseFrontmatterMarkdown(markdown);

  const expected = mdRoot([
    mdFrontmatterToml({slug, date: dates.fromISO(date)}),
    mdHeading1("Hello"),
  ]);
  expect(stripPositions(node)).toEqual(
      new PostNode(
          PostMetadata.of({slug, date: dates.fromISO(date)}),
          expected
      ));

});
