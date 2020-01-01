import { isString } from '//strings';
import * as unist from 'unist';
import * as mdast from 'mdast';

// Utilities for working with Markdown AST (mdast).

export const isCode = (n: unist.Node): n is mdast.Code => {
  const hasLang = n.lang === undefined || isString(n.lang);
  const hasMeta = n.meta === undefined || isString(n.meta);
  return n.type === 'code' && hasLang && hasMeta;
};

export const isHeading = (n: unist.Node): n is mdast.Heading => {
  const d = n.depth as number;
  return n.type === 'heading' && Number.isInteger(d) && d > 0 && d < 7;
};

export const isParagraph = (n: unist.Node): n is mdast.Paragraph => {
  return n.type === 'paragraph' && isParent(n);
};

export const isText = (n: unist.Node): n is mdast.Text => {
  return n.type === 'text' && isString(n.value);
};

export const isParent = (n: unist.Node): n is unist.Parent => {
  return Array.isArray(n.children);
};

export function checkNodeType<T extends unist.Node>(
  node: unist.Node,
  name: string,
  check: (n: unist.Node) => n is T
): asserts node is T {
  if (!check(node)) {
    const noChildren = JSON.stringify({
      ...node,
      ...{ children: '<omitted>' },
    });
    throw new Error(`Expected ${name} node but had: ${noChildren}`);
  }
}
