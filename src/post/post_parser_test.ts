import * as dates from '../dates';
import { dedent } from '../strings';
import { ZipFileEntry, Zipper } from '../zip_files';
import { PostMetadata } from './post_metadata';
import { PostNode, PostParser, TEXT_PACK_BUNDLE_PREFIX } from './post_parser';
import {
  mdHeading,
  mdPara,
  mdRoot,
  mdText,
  stripPositions,
} from './testing/markdown_nodes';

test('parses from markdown', async () => {
  const node = await PostParser.create().parseMarkdown(dedent`
    # hello
    
    \`\`\`yaml
    # Metadata
    slug: foo_bar
    date: 2019-10-08
    \`\`\`
    
    Hello world.
  `);

  const frontMatter = PostMetadata.of({
    slug: 'foo_bar',
    date: dates.fromISO('2019-10-08'),
  });
  const expected = mdRoot([
    mdHeading(1, [mdText('hello')]),
    mdPara([mdText('Hello world.')]),
  ]);
  expect(stripPositions(node)).toEqual(new PostNode(frontMatter, expected));
});

test('parses from TextPack', async () => {
  const markdown = dedent`
    # hello
    
    \`\`\`yaml
    # Metadata
    slug: foo_bar
    date: 2019-10-08
    \`\`\`
    
    Hello world.
  `;
  const buf = await Zipper.zip([
    ZipFileEntry.ofUtf8(TEXT_PACK_BUNDLE_PREFIX + '/text.md', markdown),
  ]);
  const node = await PostParser.create().parseTextPack(buf);

  const frontMatter = PostMetadata.of({
    slug: 'foo_bar',
    date: dates.fromISO('2019-10-08'),
  });
  const expected = mdRoot([
    mdHeading(1, [mdText('hello')]),
    mdPara([mdText('Hello world.')]),
  ]);
  expect(stripPositions(node)).toEqual(new PostNode(frontMatter, expected));
});
