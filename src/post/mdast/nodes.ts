import { isString } from '//strings';
import * as unist from 'unist';
import * as mdast from 'mdast';

// Utilities for working with Markdown AST (mdast).

export const isBreak = (n: unist.Node): n is mdast.Break => {
  return n.type === 'break';
};

export const isBlockquote = (n: unist.Node): n is mdast.Blockquote => {
  return n.type === 'blockquote' && isParent(n);
};

export const isCode = (n: unist.Node): n is mdast.Code => {
  const hasLang = n.lang === undefined || isString(n.lang);
  const hasMeta = n.meta === undefined || isString(n.meta);
  return n.type === 'code' && hasLang && hasMeta;
};

export const isDelete = (n: unist.Node): n is mdast.Delete => {
  return n.type === 'delete' && isParent(n);
};

export const isFootnote = (n: unist.Node): n is mdast.Footnote => {
  return n.type === 'footnote' && isParent(n);
};

export const isFootnoteDefinition = (
  n: unist.Node
): n is mdast.FootnoteDefinition => {
  return (
    n.type === 'footnoteDefinition' &&
    isParent(n) &&
    isNonEmptyString(n.identifier)
  );
};

export const isFootnoteReference = (
  n: unist.Node
): n is mdast.FootnoteReference => {
  return n.type === 'footnoteReference' && isNonEmptyString(n.identifier);
};

export const isHeading = (n: unist.Node): n is mdast.Heading => {
  const d = n.depth as number;
  return (
    n.type === 'heading' && Number.isInteger(d) && d > 0 && d < 7 && isParent(n)
  );
};

export const isEmphasis = (n: unist.Node): n is mdast.Emphasis => {
  return n.type === 'emphasis' && isParent(n);
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

export function checkType<T extends unist.Node>(
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

const isNonEmptyString = (s: any): s is string => {
  return isString(s) && s !== '';
};
