import {PostNode, PostParser} from './post_parser';

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

test('parses simple markdown', async () => {
  const vFile = await PostParser.create().parse('# hello');

  const expected = mdRoot(pos(start(1, 1, 0), end(1, 8, 7)), [
    mdHeading(1, pos(start(1, 1, 0), end(1, 8, 7), indent()), [
      mdText('hello', pos(start(1, 3, 2), end(1, 8, 7), indent())),
    ]),
  ]);
  expect(vFile).toEqual(new PostNode({}, expected));
});

test('parses front matter', async () => {
  const vFile = await PostParser.create().parse(`# hello
  
  `);

  const expected = mdRoot(pos(start(1, 1, 0), end(1, 8, 7)), [
    mdHeading(1, pos(start(1, 1, 0), end(1, 8, 7), indent()), [
      mdText('hello', pos(start(1, 3, 2), end(1, 8, 7), indent())),
    ]),
  ]);
  expect(vFile).toEqual(new PostNode({}, expected));
});
