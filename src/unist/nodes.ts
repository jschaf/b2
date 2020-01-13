import { isObject } from '//objects';
import { isString } from '//strings';
import * as unist from 'unist';
import nodeRemove from 'unist-util-remove';

export type NodeAncestors = unist.Node[];
export type NodeVisitor = (n: unist.Node, ancestors: NodeAncestors) => void;
export type NodeTest<T extends unist.Node> = (
  n: unist.Node,
  ancestors: unist.Node[]
) => n is T;

type NodeIterResult = { node: unist.Node; ancestors: NodeAncestors };

/** Generates a pre-order traversal starting at tree. */
export function* preOrderGenerator(
  tree: unist.Node
): Generator<NodeIterResult, void> {
  const ancestors: unist.Node[] = [];

  function* visit(n: unist.Node): Generator<NodeIterResult, void> {
    yield { node: n, ancestors };

    if (isParent(n)) {
      ancestors.push(n);
      for (const c of n.children) {
        for (const child of visit(c)) {
          yield { node: child.node, ancestors };
        }
      }
      ancestors.pop();
    }
  }

  for (const r of visit(tree)) {
    yield r;
  }
}

/** Applies the visitor to all nodes under n in a pre-order traversal. */
export const visitInPlace = (tree: unist.Node, visitor: NodeVisitor): void => {
  for (const { node, ancestors } of preOrderGenerator(tree)) {
    visitor(node, ancestors);
  }
};

/**
 * Returns the first node that matches the type guard from a pre-order
 * traversal.
 *
 * If no node matches, returns null.
 */
export const findNode = <T extends unist.Node>(
  tree: unist.Node,
  test: NodeTest<T>
): T | null => {
  for (const { node, ancestors } of preOrderGenerator(tree)) {
    if (test(node, ancestors)) {
      return node;
    }
  }
  return null;
};

export const removeNode = <T extends unist.Node>(
  tree: unist.Node,
  test: NodeTest<T>
): void => {
  const wrappedTest = (n: unist.Node): n is T => test(n, []);
  nodeRemove(tree, wrappedTest);
};

export const removePositionInfo = (tree: unist.Node): void => {
  visitInPlace(tree, n => delete n.position);
};

/**
 * Removes the data object from all nodes in tree.
 *
 * node.data is set by plugins for communication across different unified
 * instances.
 */
export const removeData = (tree: unist.Node): void => {
  visitInPlace(tree, n => delete n.data);
};

export const isParent = (n: unist.Node): n is unist.Parent => {
  return Array.isArray(n.children);
};

/** Type guard that returns true if n is a node. */
export const isNode = (n: unknown): n is unist.Node => {
  return isObject(n) && isString(n.type) && n.type !== '';
};

/** Type guard that returns true if ns is a node array. */
export const isNodeArray = (ns: unknown): ns is unist.Node[] => {
  if (!Array.isArray(ns)) {
    return false;
  }
  for (const n of ns) {
    if (!isNode(n)) {
      return false;
    }
  }
  return true;
};

export type Text = { type: 'text'; value: string };

export const text = (value: string): Text => ({ type: 'text', value });

export const isText = (n: unist.Node): n is Text => {
  return n.type === 'text' && isString(n.value);
};

interface NodeWithData extends unist.Node {
  data: unist.Data;
}

/** Adds data to the node if it doesn't already exist. */
export const ensureDataAttr = (n: unist.Node): NodeWithData => {
  if (!n.data) {
    n.data = {};
  }
  return n as NodeWithData;
};

export const mergeAdjacentText = (src: unist.Node[]): unist.Node[] => {
  if (src.length === 0) {
    return src;
  }

  const dest = [src[0]];
  for (let i = 1; i < src.length; i++) {
    const s = src[i];
    const d = dest[dest.length - 1];
    if (isText(s) && isText(d)) {
      d.value += s.value;
    } else {
      dest.push(s);
    }
  }
  return dest;
};
