import * as unist from 'unist';

// Shortcuts for creating HTML AST nodes (hast).
// https://github.com/syntax-tree/hastscript

export interface HtmlNode {
  type: 'element';
  children?: (HtmlNode | unist.Node)[];
  [key: string]: unknown;
}

export const htmlNode = (
  tag: string,
  props: Record<string, unknown>,
  children?: (HtmlNode | unist.Node)[]
): HtmlNode => {
  const childObj = children == null ? {} : { children };
  return { type: 'element', tagName: tag, properties: props, ...childObj };
};
