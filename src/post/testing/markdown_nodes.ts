import { checkDefined } from '//asserts';
import { removePositionInfo } from '//unist/nodes';
import * as toml from '@iarna/toml';
import { BlockContent } from 'mdast';
import { PostNode } from '../post_parser';
import * as mdast from 'mdast';

export const mdBlockquote = (
  children: mdast.BlockContent[]
): mdast.Blockquote => {
  return { type: 'blockquote', children };
};

export const mdBreak = (): mdast.Break => {
  return { type: 'break' };
};

export const mdCode = (code: string): mdast.Code => {
  return { type: 'code', value: code };
};

export const mdCodeWithLang = (lang: string, code: string): mdast.Code => {
  return { type: 'code', lang, value: code };
};

export const mdDelete = (children: mdast.PhrasingContent[]): mdast.Delete => {
  return { type: 'delete', children };
};
export const mdEmphasis = (
  children: mdast.PhrasingContent[]
): mdast.Emphasis => {
  return { type: 'emphasis', children };
};

export const mdEmphasisText = (text: string): mdast.Emphasis => {
  return mdEmphasis([mdText(text)]);
};

export const mdInlineFootnote = (
  children: mdast.PhrasingContent[]
): mdast.Footnote => {
  return { type: 'footnote', children };
};

export const mdFootnoteDef = (
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

export const mdFootnoteRef = (identifier: string): mdast.FootnoteReference => {
  return { type: 'footnoteReference', identifier, label: identifier };
};

export const mdFrontmatterToml = (value: toml.JsonMap): mdast.Content => {
  let raw = toml
    .stringify(value)
    .trimEnd()
    .replace(/T00:00:00.000Z/, '');
  // The typings for mdast don't allow anything except whitelisted types.
  // Force add toml as a supported type.
  return ({
    type: 'toml',
    value: raw,
  } as unknown) as mdast.Content;
};

export const mdHeading = (
  heading: 'h1' | 'h2' | 'h3' | 'h4' | 'h5' | 'h6',
  children: mdast.PhrasingContent[]
): mdast.Heading => {
  const match = checkDefined(
    heading.match(/h(\d)/),
    'heading regex must match'
  );
  const depth = +match[1] as 1 | 2 | 3 | 4 | 5 | 6;
  return { type: 'heading', depth: depth, children };
};

export const mdHeading1 = (child: string): mdast.Heading => {
  return { type: 'heading', depth: 1, children: [mdText(child)] };
};

export const mdListItem = (children: mdast.BlockContent[]): mdast.ListItem => {
  return {
    type: 'listItem',
    spread: false,
    // Unified parses the checked property as null but the type is boolean or
    // undefined.
    checked: (null as unknown) as boolean,
    children,
  };
};

export const mdOrderedList = (children: BlockContent[]): mdast.List => {
  return {
    type: 'list',
    ordered: true,
    spread: false,
    start: 1,
    children: children.map(c => mdListItem([c])),
  };
};

export const mdPara = (children: mdast.PhrasingContent[]): mdast.Paragraph => {
  return { type: 'paragraph', children };
};

export const mdParaText = (value: string): mdast.Paragraph => {
  return { type: 'paragraph', children: [mdText(value)] };
};

export const mdRoot = (children: mdast.Content[]): mdast.Root => {
  return { type: 'root', children };
};

export const mdText = (value: string): mdast.Text => {
  return { type: 'text', value };
};

export const stripPositions = (node: PostNode): PostNode => {
  removePositionInfo(node.node);
  return node;
};
