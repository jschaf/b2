import * as unist from 'unist';
import * as hast from 'hast-format';

// Shortcuts for creating HTML AST nodes (hast).
// https://github.com/syntax-tree/hastscript

export const hastText = (text: string): hast.Text => {
  return { type: 'text', value: text };
};

export const hastElem = (
  tagName: string,
  children: unist.Node[] = []
): hast.Element => {
  // We use the dispatcher to figure out what to render so we don't know the
  // types ahead of time.
  return { type: 'element', tagName, children: children as any };
};

export const hastElemWithProps = (
  tagName: string,
  props: hast.Properties,
  children: unist.Node[] = []
): hast.Element => {
  // We use the dispatcher to figure out what to render so we don't know the
  // types ahead of time.
  return {
    type: 'element',
    tagName,
    properties: props,
    children: children as any,
  };
};
