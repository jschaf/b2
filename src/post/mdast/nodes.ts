import { checkDefined } from '//asserts';
import { PostNode } from '//post/post_parser';
import { isString } from '//strings';
import { removePositionInfo } from '//unist/nodes';
import * as tomlLib from '@iarna/toml';
import * as mdast from 'mdast';
import { BlockContent } from 'mdast';
import * as unist from 'unist';

// Utilities for working with Markdown AST (mdast).

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
export const stripPositions = (node: PostNode): PostNode => {
  removePositionInfo(node.node);
  return node;
};
