import {
  PostNode,
  PostParser,
  TEXT_PACK_BUNDLE_PREFIX,
} from '//post/post_parser';
import {
  DEFAULT_FRONTMATTER,
  withDefaultFrontMatter,
} from '//post/testing/front_matters';
import { dedent } from '//strings';
import { ZipFileEntry, Zipper } from '//zip_files';
import {
  mdHeading,
  mdOrderedList,
  mdParaText,
  mdRoot,
  mdText,
  stripPositions,
} from './testing/markdown_nodes';

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
