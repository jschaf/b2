import * as dates from '//dates';
import { dedent } from '//strings';
import { ZipFileEntry, Zipper } from '//zip_files';
import { PostMetadata } from '//post/post_metadata';
import {
  PostNode,
  PostParser,
  TEXT_PACK_BUNDLE_PREFIX,
} from '//post/post_parser';
import {
  mdHeading,
  mdOrderedList,
  mdParaText,
  mdRoot,
  mdText,
  stripPositions,
} from './testing/markdown_nodes';

const withFrontMatter = (
  text: string,
  frontMatter: string = DEFAULT_FRONTMATTER_TEXT,
  lineNum: number = 2
): string => {
  const lines = text.split('\n');
  lines.splice(lineNum, 0, ...frontMatter.split('\n'));
  return lines.join('\n');
};

const DEFAULT_FRONTMATTER_TEXT = dedent`
    \`\`\`yaml
    # Metadata
    slug: foo_bar
    date: 2019-10-08
    \`\`\`
`;

const DEFAULT_FRONTMATTER = PostMetadata.of({
  slug: 'foo_bar',
  date: dates.fromISO('2019-10-08'),
});

test('parses from markdown', () => {
  const markdown = withFrontMatter(
    dedent`
    # hello
    
    Hello world.
  `,
    DEFAULT_FRONTMATTER_TEXT
  );
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
  const markdown = withFrontMatter(dedent`
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
  const markdown = withFrontMatter(dedent`
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
