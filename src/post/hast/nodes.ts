import { ReferenceType } from '//post/mdast/nodes';
import * as unist from 'unist';
import * as hast from 'hast-format';
import * as objects from '//objects';
import * as mdast from 'mdast';

// Shortcuts for creating HTML AST nodes (hast).
// https://github.com/syntax-tree/hastscript

/**
 * Returns the content of an image reference rather than its definition.
 *
 * Used when no definition is found matching the node's identifier.
 */
export const danglingImageRef = (n: mdast.ImageReference): hast.Text => {
  switch (n.referenceType) {
    case ReferenceType.Collapsed:
      return text(`![${n.alt}][]`);
    case ReferenceType.Full:
      return text(`![${n.alt}][${n.label || n.identifier}]`);
    default:
      throw new Error('unreachable');
  }
};

/** Creates a hast element using tagName and children. */
export const elem = (
  tagName: string,
  children: unist.Node[] = []
): hast.Element => {
  // We use the dispatcher to figure out what to compileNode so we don't know the
  // types ahead of time.
  return elemProps(tagName, {}, children);
};

/** Creates a hast element using tagName with a single text child. */
export const elemText = (tagName: string, value: string): hast.Element => {
  return elemProps(tagName, {}, [text(value)]);
};

/** Creates a hast element using tagName, props and children. */
export const elemProps = (
  tagName: string,
  props: hast.Properties,
  children: unist.Node[] = []
): hast.Element => {
  // The compiler dispatches to node compiler so we don't know the types at
  // compile time. Give up by using any.
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

export interface Raw extends unist.Literal {
  type: 'raw';
}

/** Creates a raw literal hast node. */
export const raw = (value: string): Raw => {
  return { type: 'raw', value };
};

/** Creates a text literal hast node. */
export const text = (value: string): hast.Text => {
  return { type: 'text', value };
};
