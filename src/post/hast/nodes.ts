import { isOptionalObject } from '//objects';
import { isLiteral, RefType } from '//post/mdast/nodes';
import { isString } from '//strings';
import { Comment, DocType, Element, Text } from 'hast-format';
import * as unist from 'unist';
import * as hast from 'hast-format';
import * as objects from '//objects';
import * as mdast from 'mdast';
import * as unistNodes from '//unist/nodes';

// Shortcuts for creating HTML AST nodes (hast).
// https://github.com/syntax-tree/hastscript

export interface Literal {
  value: string;
}

export interface Tag<T extends string = string> extends unist.Node {
  tagName: T;
}

export interface ParentTag<T extends string> extends Tag<T> {
  tagName: T;
  children: unist.Node[];
}

export interface LiteralTag<T extends string> extends Tag<T> {
  tagName: T;
  value: string;
}

/**
 * Returns the markdown representation of an image reference rather than its
 * definition.
 *
 * Used when no definition is found matching the node's identifier.
 *
 * https://spec.commonmark.org/0.29/#images
 */
export const danglingImageRef = (n: mdast.ImageReference): hast.Text => {
  // Prefer the label since the identifier is normalized.
  const id = n.label || n.identifier;
  switch (n.referenceType) {
    case RefType.Collapsed:
      return text(`![${id}][]`);
    case RefType.Full:
      return text(`![${n.alt}][${id}]`);
    case RefType.Shortcut:
      return text(`![${id}]`);
    default:
      throw new Error('unreachable');
  }
};

/**
 * Returns the markdown representation of a link reference rather than its
 * definition.
 *
 * Used when no definition is found matching the node's identifier.
 */
export const danglingLinkRef = (
  n: mdast.LinkReference,
  childrenCompiler: (n: mdast.LinkReference) => unist.Node[]
): unist.Node[] => {
  const children = childrenCompiler(n);
  const merge = unistNodes.mergeAdjacentText;
  // Prefer the label since the identifier is normalized.
  const id = n.label || n.identifier;
  switch (n.referenceType) {
    case RefType.Collapsed:
      return merge([text('['), ...children, text('][]')]);
    case RefType.Full:
      return merge([text('['), ...children, text(`][${id}]`)]);
    case RefType.Shortcut:
      return merge([text('['), ...children, text(']')]);
    default:
      throw new Error('unreachable');
  }
};

export const comment = (value: string): hast.Comment => {
  return { type: 'comment', value };
};

export const isComment = (n: unist.Node): n is hast.Comment => {
  return n.type === 'comment' && isString(n.value);
};

export const doctype = (): hast.DocType => {
  // Hard code HTML5 doctype.
  return { type: 'doctype', name: 'html' };
};

export const isDoctype = (n: unist.Node): n is hast.DocType => {
  return n.type === 'doctype' && isString(n.name);
};

/** Creates a hast element using tagName and children. */
export const elem = (
  tagName: string,
  children: unist.Node[] = []
): hast.Element => {
  // We use the dispatcher to figure out what to compile so we don't know the
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

export const isElem = (n: unist.Node): n is Tag => {
  return n.type === 'element' && isOptionalObject(n.properties);
};

export const isParentElem = (n: unist.Node): n is ParentTag<string> => {
  return isElem(n) && Array.isArray(n.children);
};

export const isLiteralElem = (n: unist.Node): n is LiteralTag<string> => {
  return isElem(n) && isString(n.value);
};

export const isTag = <T extends string>(
  tagName: T,
  n: unist.Node
): n is Tag<T> => {
  return isElem(n) && n.tagName === tagName;
};

export const isParentTag = <T extends string>(
  tagName: T,
  n: unist.Node
): n is ParentTag<T> => {
  return isTag(tagName, n) && Array.isArray(n.children);
};

export interface Raw extends unist.Literal {
  type: 'raw';
}

export interface Raw {
  type: 'raw';
  value: string;
}

export const isRaw = (n: unist.Node): n is Raw => {
  return n.type === 'raw' && isLiteral(n);
};

/** Creates a raw literal hast node. */
export const raw = (value: string): hast.Element => {
  return ({ type: 'raw', value } as unknown) as hast.Element;
};

export type RootContent = Element | DocType | Comment | Text;

/** Creates a raw literal hast node. */
export const root = (children: RootContent[]): hast.Root => {
  return { type: 'root', children };
};

export const isRoot = (n: unist.Node): n is hast.Root => {
  return n.type === 'root' && Array.isArray(n.children);
};

export const scriptElem = (value: string): Tag<'script'> & Literal => {
  return { type: 'element', tagName: 'script', value };
};

export const isScriptElem = (n: unist.Node): n is Tag<'script'> & Literal => {
  return n.type === 'element' && n.tagName === 'script' && isLiteral(n);
};

/** Creates a text literal hast node. */
export const text = (value: string): hast.Text => {
  return { type: 'text', value };
};

export const normalizeUri = (uri: string): string => {
  return encodeURI(uri.trim());
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
