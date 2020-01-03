import { checkDefined } from '//asserts';
import { PostNode } from '//post/post_parser';
import { isOptionalString, isString } from '//strings';
import { removePositionInfo } from '//unist/nodes';
import * as tomlLib from '@iarna/toml';
import * as mdast from 'mdast';
import { BlockContent } from 'mdast';
import * as unist from 'unist';

// Utilities for working with Markdown AST (mdast).

// break is a keyword so use lineBreak.
export const lineBreak = (): mdast.Break => {
  return { type: 'break' };
};

export const isBreak = (n: unist.Node): n is mdast.Break => {
  return n.type === 'break';
};

export const blockquote = (
  children: mdast.BlockContent[]
): mdast.Blockquote => {
  return { type: 'blockquote', children };
};

export const isBlockquote = (n: unist.Node): n is mdast.Blockquote => {
  return n.type === 'blockquote' && isParent(n);
};

export const code = (code: string): mdast.Code => {
  return { type: 'code', value: code };
};
export const codeWithLang = (lang: string, code: string): mdast.Code => {
  return { type: 'code', lang, value: code };
};

export const isCode = (n: unist.Node): n is mdast.Code => {
  const hasLang = n.lang === undefined || isString(n.lang);
  const hasMeta = n.meta === undefined || isString(n.meta);
  return n.type === 'code' && hasLang && hasMeta;
};

type DefinitionProps = { label?: string; title?: string };

export const definition = (id: string, url: string): mdast.Definition => {
  return definitionProps(id, url, {});
};

export const definitionProps = (
  id: string,
  url: string,
  props: DefinitionProps
): mdast.Definition => {
  return { type: 'definition', identifier: id, url, ...props };
};

export const isDefinition = (n: unist.Node): n is mdast.Definition => {
  return n.type === 'definition' && isAssociation(n) && isResource(n);
};

// delete is a keyword so use deleted.
export const deleted = (children: mdast.PhrasingContent[]): mdast.Delete => {
  return { type: 'delete', children };
};

export const isDelete = (n: unist.Node): n is mdast.Delete => {
  return n.type === 'delete' && isParent(n);
};

export const emphasis = (children: mdast.PhrasingContent[]): mdast.Emphasis => {
  return { type: 'emphasis', children };
};

export const emphasisText = (s: string): mdast.Emphasis => {
  return emphasis([text(s)]);
};

export const isEmphasis = (n: unist.Node): n is mdast.Emphasis => {
  return n.type === 'emphasis' && isParent(n);
};

export const footnote = (children: mdast.PhrasingContent[]): mdast.Footnote => {
  return { type: 'footnote', children };
};

export const isFootnote = (n: unist.Node): n is mdast.Footnote => {
  return n.type === 'footnote' && isParent(n);
};

export const footnoteDef = (
  identifier: string,
  children: mdast.BlockContent[]
): mdast.FootnoteDefinition => {
  return {
    type: 'footnoteDefinition',
    identifier,
    label: identifier,
    children,
  };
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

export const footnoteRef = (identifier: string): mdast.FootnoteReference => {
  return { type: 'footnoteReference', identifier, label: identifier };
};

export const isFootnoteReference = (
  n: unist.Node
): n is mdast.FootnoteReference => {
  return n.type === 'footnoteReference' && isNonEmptyString(n.identifier);
};

type HeadingLevel = 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6';

export const heading = (
  h: HeadingLevel,
  children: mdast.PhrasingContent[]
): mdast.Heading => {
  const match = checkDefined(h.match(/h(\d)/), 'heading regex must match');
  const depth = +match[1] as 1 | 2 | 3 | 4 | 5 | 6;
  return { type: 'heading', depth: depth, children };
};

export const headingText = (h: HeadingLevel, child: string): mdast.Heading => {
  return heading(h, [text(child)]);
};

export const isHeading = (n: unist.Node): n is mdast.Heading => {
  const d = n.depth as number;
  return (
    n.type === 'heading' && Number.isInteger(d) && d > 0 && d < 7 && isParent(n)
  );
};

export const html = (value: string): mdast.HTML => {
  return { type: 'html', value };
};

export const isHTML = (n: unist.Node): n is mdast.HTML => {
  return n.type === 'html' && isLiteral(n);
};

type ImageProps = Omit<mdast.Resource, 'url'> & mdast.Alternative;

export const image = (url: string): mdast.Image => {
  return imageProps(url, {});
};

export const imageProps = (url: string, props: ImageProps): mdast.Image => {
  return { type: 'image', url, ...props };
};

export const isImage = (n: unist.Node): n is mdast.Image => {
  return n.type === 'image' && isResource(n) && isAlternative(n);
};

export type ImageRefProps = { label?: string; alt?: string };

export const imageRef = (id: string, ref: RefType): mdast.ImageReference => {
  return imageRefProps(id, ref, {});
};

export const imageRefProps = (
  id: string,
  ref: RefType,
  props: ImageRefProps
): mdast.ImageReference => {
  return {
    type: 'imageReference',
    identifier: id,
    referenceType: ref,
    ...props,
  };
};

export const isImageRef = (n: unist.Node): n is mdast.ImageReference => {
  return n.type === 'imageReference' && isReference(n) && isAlternative(n);
};

export const inlineCode = (value: string): mdast.InlineCode => {
  return { type: 'inlineCode', value: value };
};

export const isInlineCode = (n: unist.Node): n is mdast.InlineCode => {
  return n.type === 'inlineCode' && isLiteral(n);
};

export const link = (
  url: string,
  children: mdast.StaticPhrasingContent[]
): mdast.Link => {
  return linkProps(url, {}, children);
};

export type LinkProps = { title?: string };

export const linkProps = (
  url: string,
  props: LinkProps,
  children: mdast.StaticPhrasingContent[]
): mdast.Link => {
  return { type: 'link', url, ...props, children };
};

export const linkText = (url: string, value: string): mdast.Link => {
  return link(url, [text(value)]);
};

export type LinkRefProps = { label?: string };
export const isLink = (n: unist.Node): n is mdast.Link => {
  return n.type === 'link' && isResource(n);
};

export const linkRef = (
  id: string,
  refType: RefType,
  children: mdast.StaticPhrasingContent[]
): mdast.LinkReference => {
  return linkRefProps(id, refType, {}, children);
};

export const linkRefProps = (
  id: string,
  refType: RefType,
  props: LinkRefProps,
  children: mdast.StaticPhrasingContent[]
): mdast.LinkReference => {
  return {
    type: 'linkReference',
    identifier: id,
    referenceType: refType,
    ...props,
    children,
  };
};

export const linkRefText = (
  id: string,
  refType: RefType,
  value: string
): mdast.LinkReference => {
  return linkRef(id, refType, [text(value)]);
};

export const isLinkRef = (n: unist.Node): n is mdast.LinkReference => {
  return n.type === 'linkReference' && isParent(n) && isReference(n);
};

export const listItem = (children: mdast.BlockContent[]): mdast.ListItem => {
  return {
    type: 'listItem',
    spread: false,
    // Unified parses the checked property as null but the type is boolean or
    // undefined.
    checked: (null as unknown) as boolean,
    children,
  };
};

export const orderedList = (children: BlockContent[]): mdast.List => {
  return {
    type: 'list',
    ordered: true,
    spread: false,
    start: 1,
    children: children.map(c => listItem([c])),
  };
};

export const paragraph = (
  children: mdast.PhrasingContent[]
): mdast.Paragraph => {
  return { type: 'paragraph', children };
};

export const paragraphText = (value: string): mdast.Paragraph => {
  return paragraph([text(value)]);
};

export const isParagraph = (n: unist.Node): n is mdast.Paragraph => {
  return n.type === 'paragraph' && isParent(n);
};

export const root = (children: mdast.Content[]): mdast.Root => {
  return { type: 'root', children };
};

export const isRoot = (n: unist.Node): n is mdast.Root => {
  return n.type === 'root' && isParent(n);
};

export const strong = (children: mdast.PhrasingContent[]): mdast.Strong => {
  return { type: 'strong', children };
};

export const strongText = (value: string): mdast.Strong => {
  return strong([text(value)]);
};

export const isStrong = (n: unist.Node): n is mdast.Strong => {
  return n.type === 'strong' && isParent(n);
};

export const text = (value: string): mdast.Text => {
  return { type: 'text', value };
};

export const isText = (n: unist.Node): n is mdast.Text => {
  return n.type === 'text' && isString(n.value);
};

interface Toml extends mdast.Literal {
  type: 'toml';
}

export const toml = (map: tomlLib.JsonMap): BlockContent => {
  let raw = tomlLib.stringify(map).trimEnd();
  // The mdast typings only allow known types.
  return ({
    type: 'toml',
    value: raw,
  } as unknown) as mdast.BlockContent;
};

export const tomlFrontmatter = (map: tomlLib.JsonMap): mdast.BlockContent => {
  const t = toml(map);
  checkType(t, 'toml', isToml);
  // Normalize date.
  t.value = t.value.replace(/T00:00:00.000Z/, '');
  return (t as unknown) as mdast.BlockContent;
};

export const isToml = (n: unist.Node): n is Toml => {
  return n.type === 'toml' && isString(n.value);
};

export const isLiteral = (n: unist.Node): n is unist.Literal => {
  return n.value && isString(n.value);
};

export const isParent = (n: unist.Node): n is unist.Parent => {
  return Array.isArray(n.children);
};

/**
 * RefType represents a marker that is associated to another node.
 *
 * https://github.com/syntax-tree/mdast#reference
 */
export const enum RefType {
  /** The reference is implicit, its identifier inferred from its content. */
  Shortcut = 'shortcut',
  /** The reference is explicit, its identifier inferred from its content. */
  Collapsed = 'collapsed',
  /** The reference is explicit, its identifier explicitly set. */
  Full = 'full',
}

// Combine Node and an mdast mixin. Necessary since the guards below take
// a unist.Node but return a guard on the mixin. This type makes the mixin also
// a node.
type WithNode<T> = unist.Node & T;

export const isAlternative = (
  n: unist.Node
): n is WithNode<mdast.Alternative> => {
  return isOptionalString(n.alt);
};

export const isAssociation = (
  n: unist.Node
): n is WithNode<mdast.Association> => {
  return isNonEmptyString(n.identifier) && isOptionalString(n.label);
};

export const isReference = (n: unist.Node): n is WithNode<mdast.Reference> => {
  let rt = n.referenceType;
  const isValidRef =
    rt === RefType.Shortcut || rt === RefType.Collapsed || rt === RefType.Full;
  return isValidRef && isAssociation(n);
};

export const isResource = (n: unist.Node): n is WithNode<mdast.Resource> => {
  const isValidTitle = n.title ? isNonEmptyString(n.title) : true;
  return n.url && isNonEmptyString(n.url) && isValidTitle;
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

export const stripPositions = (node: PostNode): PostNode => {
  removePositionInfo(node.node);
  return node;
};

/**
 * Normalizes a link label according to commonmark.
 *
 * https://spec.commonmark.org/0.29/#matches
 */
export const normalizeLabel = (l: string): string => {
  // Perform Unicode case fold.
  const lowered = l.toLowerCase();
  // Strip leading and trailing whitespace.
  const trimmed = lowered.trim();
  // Collapse consecutive internal whitespace to a single space.
  return trimmed.replace(/\s+/g, ' ');
};
