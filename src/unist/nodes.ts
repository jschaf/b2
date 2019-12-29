import {isObject} from '//objects';
import {isString} from '//strings';
import * as unist from 'unist';

export type NodeAncestors = unist.Node[];
export type NodeVisitor = (n: unist.Node, ancestors: unist.Node[]) => void
export type NodeTest<T extends unist.Node> = (n: unist.Node, ancestors: unist.Node[]) => n is T;

/** Applies the visitor to all nodes under n in a pre-order traversal. */
export const visitInPlace = (node: unist.Node, visitor: NodeVisitor): void => {
  const ancestors: unist.Node[] = [];

  const visit = (n: unist.Node) => {
    visitor(n, ancestors);

    if (hasChildren(n)) {
      ancestors.push(n);
      for (const c of n.children) {
        visit(c);
      }
      ancestors.pop();
    }
  };

  visit(node);
};

export const findNode = <T extends unist.Node>(
    tree: unist.Node, test: NodeTest<T>,
    ): T | null => {
  const ancestors: unist.Node[] = [];

  const visit = (n: unist.Node): T | null => {
    if (test(n, ancestors)) {
      return n;
    }

    if (hasChildren(n)) {
      ancestors.push(n);
      for (const c of n.children) {
        const r = visit(c);
        if (r !== null) {
          return r;
        }
      }
      ancestors.pop();
    }
    return null;
  };

  return visit(tree);
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
