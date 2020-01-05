import * as md from '//post/mdast/nodes';
import { PostMetadata } from '//post/metadata';
import { PostNode, PostParser, TEXT_PACK_BUNDLE_PREFIX } from '//post/parser';
import {
  DEFAULT_FRONTMATTER,
  withDefaultFrontMatter,
} from '//post/testing/front_matters';
import { dedent } from '//strings';
import { ZipFileEntry, Zipper } from '//zip_files';
import * as dates from '//dates';

test('parses from markdown', () => {
  const markdown = withDefaultFrontMatter(dedent`
    # hello
    
    Hello world.
   `);
  const node = PostParser.create().parseMarkdown(markdown);

  const expected = md.root([
    md.heading('h1', [md.text('hello')]),
    md.paragraphText('Hello world.'),
  ]);
  expect(md.stripPositions(node)).toEqual(
    new PostNode(DEFAULT_FRONTMATTER, expected)
  );
});

test('parses paragraph followed immediately by a list', () => {
  const markdown = withDefaultFrontMatter(dedent`
    Hello world.
    1. md.text
  `);
  const node = PostParser.create().parseMarkdown(markdown);

  const expected = md.root([
    md.paragraphText('Hello world.'),
    md.orderedList([md.paragraphText('md.text')]),
  ]);
  expect(md.stripPositions(node).node).toEqual(expected);
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

  const expected = md.root([
    md.heading('h1', [md.text('hello')]),
    md.paragraphText('Hello world.'),
  ]);
  expect(md.stripPositions(node)).toEqual(
    new PostNode(DEFAULT_FRONTMATTER, expected)
  );
});

test('parses from frontmatter markdown', async () => {
  const slug = 'foo_qux';
  const date = '2019-10-17';
  const markdown = dedent`
    +++
    slug = "${slug}"
    date = ${date}
    +++
    
    # Hello
  `;

  const node = PostParser.create().parseMarkdown(markdown);

  const expected = md.root([
    md.tomlFrontmatter({ slug, date: dates.fromISO(date) }),
    md.headingText('h1', 'Hello'),
  ]);
  expect(md.stripPositions(node)).toEqual(
    new PostNode(
      PostMetadata.parse({ slug, date: dates.fromISO(date) }),
      expected
    )
  );
});
