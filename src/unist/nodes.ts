import {isObject} from '//objects';
import {isString} from '//strings';
import * as unist from 'unist';

export type NodeAncestors = unist.Node[];
export type NodeVisitor = (n: unist.Node, ancestors: NodeAncestors) => void
export type NodeTest<T extends unist.Node> = (n: unist.Node, ancestors: unist.Node[]) => n is T;

type NodeIterResult = {node: unist.Node, ancestors: NodeAncestors}

/** Generates a pre-order traversal starting at tree. */
export function* preOrderGenerator(tree: unist.Node): Generator<NodeIterResult, void> {
  const ancestors: unist.Node[] = [];

  function* visit(n: unist.Node): Generator<NodeIterResult, void> {
    yield {node: n, ancestors};

    if (hasChildren(n)) {
      ancestors.push(n);
      for (const c of n.children) {
        for (const child of visit(c)) {
          yield {node: child.node, ancestors}
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
  for (const {node, ancestors} of preOrderGenerator(tree)) {
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
    tree: unist.Node, test: NodeTest<T>,
    ): T | null => {
  for (const {node, ancestors} of preOrderGenerator(tree)) {
    if (test(node, ancestors)) {
      return node;
    }
  }
  return null;
};

export const removePositionInfo = (tree: unist.Node): void => {
  visitInPlace(tree, n => delete n.position);
};

/** Type guard that returns true if a node has node children. */
export const hasChildren =
    (n: unist.Node): n is unist.Node & { children: unist.Node[] } => {
      if (!n.children) {
        return false;
      }

      if (!Array.isArray(n.children)) {
        return false;
      }

      for (const c of n.children) {
        if (!isNode(c)) {
          return false;
        }
      }

      return true;
    };

/** Type guard that returns true if n is a node. */
export const isNode = (n: unknown): n is unist.Node => {
  return isObject(n) && isString(n.type) && n.type !== '';
};
