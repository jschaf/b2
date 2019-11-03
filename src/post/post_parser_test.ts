import { PostNode, PostParser, TEXT_PACK_BUNDLE_PREFIX } from './post_parser';
import { dedent } from '../strings';
import { PostMetadata } from './post_metadata';
import * as dates from '../dates';
import removePosition from 'unist-util-remove-position';
import { ZipFileEntry, Zipper } from '../zip_files';

type MdNode = { type: string; children?: MdNode[] };

const mdNode = (type: string, params: Object, children?: MdNode[]): MdNode => {
  const childObj = children == null ? {} : { children };
  return { type, ...params, ...childObj };
};

const mdRoot = (children: MdNode[]): MdNode => {
  return mdNode('root', {}, children);
};

const mdHeading = (depth: number, children: MdNode[]): MdNode => {
  return mdNode('heading', { depth }, children);
};

const mdText = (value: string): MdNode => {
  return mdNode('text', { value });
};

const mdPara = (children: MdNode[]): MdNode => {
  return mdNode('paragraph', {}, children);
};

const stripPositions = (node: PostNode): PostNode => {
  const forceDelete = true;
  return new PostNode(node.metadata, removePosition(node.node, forceDelete));
};

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
