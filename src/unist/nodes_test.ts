import { lossyClone } from '//objects';
import * as md from '//post/mdast/nodes';
import {
  findNode,
  isText,
  mergeAdjacentText,
  NodeVisitor,
  preOrderGenerator,
  text,
  visitInPlace,
} from '//unist/nodes';
import * as unist from 'unist';

type Ancestors = unist.Node[];

describe('preOrderGenerator', () => {
  it('should iterate in preOrder', () => {
    const nodes: [unist.Node, Ancestors][] = [];
    const n1a = md.text('1 left');
    const n1b = md.text('1 mid');
    const n1 = md.paragraph([n1a, n1b]);
    const n2a = md.text('2 left');
    const n2 = md.paragraph([n2a]);
    const n0 = md.root([n1, n2]);

    for (const { node, ancestors } of preOrderGenerator(n0)) {
      nodes.push([node, ancestors.map(lossyClone)]);
    }

    expect(nodes).toEqual([
      [n0, []],
      [n1, [n0]],
      [n1a, [n0, n1]],
      [n1b, [n0, n1]],
      [n2, [n0]],
      [n2a, [n0, n2]],
    ]);
  });
});

describe('visitInPlace', () => {
  it('should visit in-order', () => {
    const nodes: [unist.Node, Ancestors][] = [];
    const visitor: NodeVisitor = (n, ancestors) => {
      nodes.push([n, ancestors.map(a => lossyClone(a))]);
    };
    const n1a = md.text('1 left');
    const n1b = md.text('1 mid');
    const n1 = md.paragraph([n1a, n1b]);
    const n2a = md.text('2 left');
    const n2 = md.paragraph([n2a]);
    const n0 = md.root([n1, n2]);

    visitInPlace(n0, visitor);

    expect(nodes).toEqual([
      [n0, []],
      [n1, [n0]],
      [n1a, [n0, n1]],
      [n1b, [n0, n1]],
      [n2, [n0]],
      [n2a, [n0, n2]],
    ]);
  });
});

describe('findNode', () => {
  const n1a = md.text('1 left');
  const n1b = md.text('1 mid');
  const n1 = md.paragraph([n1a, n1b]);
  const n2a = md.text('2 left');
  const n2 = md.paragraph([n2a]);
  const n0 = md.root([n1, n2]);

  it('should find the root node', () => {
    let isRoot = (n: unist.Node): n is { type: 'root' } => n.type === 'root';

    expect(findNode(n0, isRoot)).toEqual(n0);
  });

  it('should find the first paragraph node', () => {
    let isPara = (n: unist.Node): n is { type: 'paragraph' } =>
      n.type === 'paragraph';

    expect(findNode(n0, isPara)).toEqual(n1);
  });
});

describe('mergeAdjacentText', () => {
  const o = { type: 'o' };
  const t = text;
  const data: [string, unist.Node[], unist.Node[]][] = [
    ['empty', [], []],
    ['1 text', [t('foo')], [t('foo')]],
    ['1 other', [o], [o]],
    ['2 text', [t('foo'), t('bar')], [t('foobar')]],
    ['3 text', [t('a'), t('b'), t('c')], [t('abc')]],
    ['split text', [t('a'), t('b'), o, t('c'), t('d')], [t('ab'), o, t('cd')]],
  ];
  const fmt = (ns: unist.Node[]): string =>
    '[' +
    ns.map(n => (isText(n) ? n.value : `{type=${n.type}}`)).join(',') +
    ']';
  for (const [name, input, expected] of data) {
    it(`${name}: input=${fmt(input)}, expected=${fmt(expected)}`, () => {
      const nodes = mergeAdjacentText(input);
      expect(nodes).toEqual(expected);
    });
  }
});
