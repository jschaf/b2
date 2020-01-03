import * as unist from 'unist';
import * as hast from 'hast-format';
import * as objects from '//objects';

// Shortcuts for creating HTML AST nodes (hast).
// https://github.com/syntax-tree/hastscript

export interface Raw extends unist.Literal {
  type: 'raw';
}

export const hastRaw = (text: string): Raw => {
  return { type: 'raw', value: text };
};

export const hastText = (text: string): hast.Text => {
  return { type: 'text', value: text };
};

export const hastElem = (
  tagName: string,
  children: unist.Node[] = []
): hast.Element => {
  // We use the dispatcher to figure out what to compileNode so we don't know the
  // types ahead of time.
  return hastElemWithProps(tagName, {}, children);
};

export const hastElemText = (tagName: string, text: string): hast.Element => {
  return hastElemWithProps(tagName, {}, [hastText(text)]);
};

export const hastElemWithProps = (
  tagName: string,
  props: hast.Properties,
  children: unist.Node[] = []
): hast.Element => {
  // We use the dispatcher to figure out what to compileNode so we don't know the
  // types of the children ahead of time so use any.
  const base: hast.Element = {
    type: 'element',
    tagName,
    children: children as any,
  };
  if (!objects.isEmpty(props)) {
    base.properties = props;
  }
  return base;
};
