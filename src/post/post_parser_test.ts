import {PostNode, PostParser} from './post_parser';
import {dedent} from "../strings";
import {PostMetadata} from "./post_metadata";
import * as dates from '../dates';

type Loc = { line: number, column: number, offset: number };
type Pos = { position: { start: Loc, end: Loc, indent?: any[] } };
type MdNode = { type: string, children?: MdNode[] } & Pos;

const start = (line: number, column: number, offset: number): { start: Loc } => {
  return {start: {line, column, offset}};
};

const end = (line: number, column: number, offset: number): { end: Loc } => {
  return {end: {line, column, offset}};
};

const indent = (): { indent: any[] } => {
  return {indent: []};
};

const pos = (start: { start: Loc }, end: { end: Loc }, indent?: { indent: any[] }): Pos => {
  return {position: {...start, ...end, ...indent}};
};

const mdNode = (
    type: string, pos: Pos, params: Object, children?: MdNode[]
): MdNode => {
  const childObj = children == null ? {} : {children};
  return {type, ...pos, ...params, ...childObj};
};

const mdRoot = (pos: Pos, children: MdNode[]): MdNode => {
  return mdNode('root', pos, {}, children);
};

const mdHeading = (depth: number, pos: Pos, children: MdNode[]): MdNode => {
  return mdNode('heading', pos, {depth}, children);
};

const mdText = (value: string, pos: Pos): MdNode => {
  return mdNode('text', pos, {value});
};

test('parses front matter', async () => {
  const vFile = await PostParser.create().parse(dedent`
    # hello
    
    \`\`\`yaml
    '# Metadata',
    'slug: foo_bar',
    'date: 2019-10-08',
    \`\`\`
  `);

  const frontMatter = PostMetadata.of({
    slug: 'foo_bar',
    date: dates.fromISO('2019-10-08')
  });
  const expected = mdRoot(pos(start(1, 1, 0), end(1, 8, 7)), [
    mdHeading(1, pos(start(1, 1, 0), end(1, 8, 7), indent()), [
      mdText('hello', pos(start(1, 3, 2), end(1, 8, 7), indent())),
    ]),
  ]);
  expect(vFile).toEqual(new PostNode(frontMatter, expected));
});
