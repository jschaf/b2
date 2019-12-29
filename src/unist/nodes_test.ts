import {lossyClone} from '//objects';
import {mdPara, mdRoot, mdText} from '//post/testing/markdown_nodes';
import {findNode, NodeVisitor, visitInPlace} from '//unist/nodes';
import * as unist from 'unist';

type Ancestors = unist.Node[];
describe('visitInPlace', () => {
  it('should visit in-order', () => {
    const nodes: [unist.Node, Ancestors][] = [];
    const visitor: NodeVisitor = (n, ancestors) => {
      nodes.push([n, ancestors.map(a => lossyClone(a))]);
    };
    const n1a = mdText('1 left');
    const n1b = mdText('1 mid');
    const n1 = mdPara([n1a, n1b]);
    const n2a = mdText('2 left');
    const n2 = mdPara([n2a]);
    const n0 = mdRoot([n1, n2]);

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
  const n1a = mdText('1 left');
  const n1b = mdText('1 mid');
  const n1 = mdPara([n1a, n1b]);
  const n2a = mdText('2 left');
  const n2 = mdPara([n2a]);
  const n0 = mdRoot([n1, n2]);

  it('should find the root node', () => {
    let isRoot = (n: unist.Node): n is { type: 'root' } => n.type === 'root';

    expect(findNode(n0, isRoot)).toEqual(n0);
  });

  it('should find the first para node', () => {
    let isPara = (n: unist.Node): n is { type: 'paragraph' } =>
        n.type === 'paragraph';

    expect(findNode(n0, isPara)).toEqual(n1);
  });
});
