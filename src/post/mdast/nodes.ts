import { checkDefined } from '//asserts';
import { isOptionalBoolean } from '//booleans';
import { isOptionalNumber } from '//numbers';
import { isOptionalString, isString } from '//strings';
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
  // The mdast parser uses null for meta instead of undefined.
  const hasMeta = isOptionalString(n.meta) || n.meta === null;
  return (
    n.type === 'code' && isLiteral(n) && isOptionalString(n.lang) && hasMeta
  );
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

type ListProps = { ordered?: boolean; start?: number; spread?: boolean };

type ShortcutListContent =
  | mdast.ListContent
  | mdast.BlockContent
  | mdast.BlockContent[];

export const listProps = (
  props: ListProps,
  children: ShortcutListContent[]
): mdast.List => {
  const items: mdast.ListItem[] = [];
  for (const child of children) {
    if (Array.isArray(child)) {
      items.push(listItem(child));
    } else if (isListItem(child)) {
      items.push(child);
    } else if (isBlockContent(child)) {
      items.push(listItem([child]));
    } else {
      throw new Error('unknown list item shortcut');
    }
  }
  return { type: 'list', ...props, children: items };
};

export const list = (children: ShortcutListContent[]): mdast.List => {
  return listProps({}, children);
};

export const listText = (items: string[]): mdast.List => {
  const listItems = items.map(i => listItem([paragraphText(i)]));
  return list(listItems);
};

export const isList = (n: unist.Node): n is mdast.List => {
  // The mdast parser uses null for start instead of undefined.
  const isValidStart = isOptionalNumber(n.start) || n.start === null;
  return (
    n.type === 'list' &&
    isParent(n) &&
    isOptionalBoolean(n.ordered) &&
    isOptionalBoolean(n.spread) &&
    isValidStart
  );
};

export type ListItemProps = { checked?: boolean; spread?: boolean };

export const listItemProps = (
  props: ListItemProps,
  children: mdast.BlockContent[]
): mdast.ListItem => {
  const li: mdast.ListItem = {
    type: 'listItem',
    spread: false,
    ...props,
    children,
  };
  if (li.checked === undefined) {
    // Unified parses the checked property as null but the type is boolean or
    // undefined.
    li.checked = (null as unknown) as boolean;
  }
  return li;
};

export const listItem = (children: mdast.BlockContent[]): mdast.ListItem => {
  return listItemProps({}, children);
};

export const listItemText = (value: string): mdast.ListItem => {
  return listItem([paragraphText(value)]);
};

export const isListItem = (n: unist.Node): n is mdast.ListItem => {
  // The mdast parser uses null for checked instead of undefined.
  const isCheckedValid = isOptionalBoolean(n.checked) || n.checked === null;
  return (
    n.type === 'listItem' &&
    isParent(n) &&
    isCheckedValid &&
    isOptionalBoolean(n.spread)
  );
};

export const orderedList = (children: ShortcutListContent[]): mdast.List => {
  return listProps({ ordered: true, spread: false, start: 1 }, children);
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

export type TableProps = { align?: mdast.AlignType[] };

export const tableProps = (
  props: TableProps,
  children: mdast.TableContent[]
): mdast.Table => {
  return { type: 'table', ...props, children };
};

export type TableShortcutRow =
  | mdast.TableRow
  | mdast.PhrasingContent[][]
  | mdast.TableCell[]
  | mdast.PhrasingContent[];

/**
 * Creates an mdast table.
 *
 * Allows omitting intermediate tableRow and tableCell nodes since they can
 * usually be inferred. The rows in the example below each produce identical
 * rows.
 *
 *     table([
 *         [text('a'), [text('b'), text('c')]],
 *         [text('a'), tableCell([text('b'), text('c')])],
 *         tableRow([tableCellText('a'), tableCell([text('b'), text('c')])]),
 *     ])
 */
export const table = (children: TableShortcutRow[]): mdast.Table => {
  const rows: mdast.TableRow[] = [];
  for (const child of children) {
    if (Array.isArray(child)) {
      // A shortcut for a row of an array of content.
      const rowContent: mdast.TableCell[] = [];
      for (const cellOrContent of child) {
        if (Array.isArray(cellOrContent)) {
          // A shortcut for a table cell with an array of children.
          rowContent.push(tableCell(cellOrContent));
        } else if (isTableCell(cellOrContent)) {
          rowContent.push(cellOrContent);
        } else if (isPhrasingContent(cellOrContent)) {
          // A shortcut for a table cell with a single child.
          rowContent.push(tableCell([cellOrContent]));
        } else {
          throw new Error('unknown node for building a table cell');
        }
      }
      rows.push(tableRow(rowContent));
    } else if (isTableRow(child)) {
      rows.push(child);
    } else {
      throw new Error('unknown node for building a table row');
    }
  }
  if (rows.length > 0) {
    const numCells = rows[0].children.length;
    for (const row of rows) {
      if (row.children.length !== numCells) {
        const r = JSON.stringify(row.children);
        throw new Error(
          `Uneven table, 1st row had ${numCells} cells but ` +
            `found row with ${row.children.length} cells: ${r}`
        );
      }
    }
  }
  return tableProps({}, rows);
};

export const isTable = (n: unist.Node): n is mdast.Table => {
  return n.type === 'table' && isParent(n);
};

export const tableRow = (children: mdast.RowContent[]): mdast.TableRow => {
  return { type: 'tableRow', children };
};

export const isTableRow = (n: unist.Node): n is mdast.TableRow => {
  return n.type === 'tableRow' && isParent(n);
};

export const tableCell = (
  children: mdast.PhrasingContent[]
): mdast.TableCell => {
  return { type: 'tableCell', children };
};

export const tableCellText = (value: string): mdast.TableCell => {
  return tableCell([text(value)]);
};

export const isTableCell = (n: unist.Node): n is mdast.TableCell => {
  return n.type === 'tableCell' && isParent(n);
};

export const text = (value: string): mdast.Text => {
  return { type: 'text', value };
};

export const isText = (n: unist.Node): n is mdast.Text => {
  return n.type === 'text' && isString(n.value);
};

export const thematicBreak = (): mdast.ThematicBreak => {
  return { type: 'thematicBreak' };
};

export const isThematicBreak = (n: unist.Node): n is mdast.ThematicBreak => {
  return n.type === 'thematicBreak';
};

interface Toml extends mdast.Literal {
  type: 'toml';
}

export const toml = (map: tomlLib.JsonMap): BlockContent => {
  let raw = tomlLib.stringify(map).trimEnd();
  return tomlText(raw);
};

export const tomlText = (value: string): BlockContent => {
  return ({
    type: 'toml',
    value,
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
  return n.type === 'toml' && isLiteral(n);
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

export const isBlockContent = (n: unist.Node): n is mdast.BlockContent => {
  return (
    isParagraph(n) ||
    isHeading(n) ||
    isThematicBreak(n) ||
    isBlockquote(n) ||
    isList(n) ||
    isTable(n) ||
    isHTML(n) ||
    isCode(n)
  );
};

export const isPhrasingContent = (
  n: unist.Node
): n is mdast.PhrasingContent => {
  return isLink(n) || isLinkRef(n) || isStaticPhrasingContent(n);
};

export const isStaticPhrasingContent = (
  n: unist.Node
): n is mdast.StaticPhrasingContent => {
  return (
    isText(n) ||
    isEmphasis(n) ||
    isStrong(n) ||
    isDelete(n) ||
    isHTML(n) ||
    isInlineCode(n) ||
    isBreak(n) ||
    isImage(n) ||
    isImageRef(n) ||
    isFootnote(n) ||
    isFootnoteReference(n)
  );
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

const isNonEmptyString = (s: unknown): s is string => {
  return isString(s) && s !== '';
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
