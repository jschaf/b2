import { checkDefined, checkState } from '//asserts';
import { PostMetadata } from '//post/metadata';
import * as unistNodes from '//unist/nodes';
import * as md from '//post/mdast/nodes';
import * as mdast from 'mdast';
import * as unist from 'unist';

/**
 * A wrapper around an mdast root node to hold state needed to compileNode the mdast
 * tree to hast.
 */
export class PostAST {
  static readonly INLINE_FOOTNOTE_DATA_KEY = 'generated_footnote_id';

  // Visible for testing.
  readonly defsById: Map<string, mdast.Definition> = new Map();
  readonly fnDefsById: Map<string, mdast.FootnoteDefinition> = new Map();

  // TODO: use actual vfile.
  private constructor(
    readonly metadata: PostMetadata,
    readonly mdastNode: unist.Node
  ) {}

  static fromMdast(n: unist.Node): PostAST {
    const meta = PostMetadata.parseFromMdast(n) || PostMetadata.empty();
    const tree = PostMetadata.normalizeMdast(n);
    const ast = new PostAST(meta, tree);
    addAllDefs(ast);
    addAllFnDefs(ast);
    return ast;
  }

  private static inlineFootnotePrefix = 'gen-';

  // visible for testing.
  static newInlineFootnoteId(id: number): string {
    return PostAST.inlineFootnotePrefix + id;
  }

  addDefinition(def: mdast.Definition): void {
    const id = def.identifier;
    checkDefined(id, 'Definition must have an id');
    checkState(id !== '', 'Definition id must not be empty');
    // The commonmark spec says labels are case-insensitive.
    // https://spec.commonmark.org/0.29/#matches.
    const normalizedId = md.normalizeLabel(id);
    if (this.defsById.has(normalizedId)) {
      // Commonmark says the first definition takes precedence. Error since we
      // don't want to support that. Also applies in addFootnoteDef.
      // https://spec.commonmark.org/0.29/#example-173
      throw new Error(
        `Duplicate definition id=${id}, normalized=${normalizedId}`
      );
    }
    this.defsById.set(normalizedId, def);
  }

  getDefinition(id: string): mdast.Definition | null {
    const nId = md.normalizeLabel(id);
    return this.defsById.get(nId) || null;
  }

  addFootnoteDef(def: mdast.FootnoteDefinition): void {
    const id = def.identifier;
    checkDefined(id, 'Footnote definition must have an identifier');
    checkState(id !== '', 'Footnote definition identifier must not be empty');
    // The commonmark spec says labels are case-insensitive.
    // https://spec.commonmark.org/0.29/#matches.
    const normalizedId = md.normalizeLabel(id);
    checkState(
      !normalizedId.startsWith(PostAST.inlineFootnotePrefix),
      `Footnote definition id=${normalizedId} must not start ` +
        `with ${PostAST.inlineFootnotePrefix}`
    );
    if (this.fnDefsById.has(normalizedId)) {
      throw new Error(
        `Duplicate footnote definition id=${id}, normalized=${normalizedId}`
      );
    }
    this.fnDefsById.set(normalizedId, def);
  }

  addInlineFootnoteDef(nextSeq: number, def: mdast.Footnote): void {
    // Skipping normalization since we generate already normalized labels.
    const fnId = PostAST.inlineFootnotePrefix + nextSeq;
    // Store the ID so we can render the footnote reference.
    unistNodes.ensureDataAttr(def).data[
      PostAST.INLINE_FOOTNOTE_DATA_KEY
    ] = fnId;
    const fnDef = md.footnoteDef(fnId, [md.paragraph(def.children)]);
    this.fnDefsById.set(fnId, fnDef);
  }

  getFootnoteDef(id: string): mdast.FootnoteDefinition | null {
    const nId = md.normalizeLabel(id);
    return this.fnDefsById.get(nId) || null;
  }
}

const addAllDefs = (p: PostAST): void => {
  for (const { node } of unistNodes.preOrderGenerator(p.mdastNode)) {
    if (!md.isDefinition(node)) {
      continue;
    }
    p.addDefinition(node);
  }
};

/**
 * Extract all footnote definitions to a map of the footnote identifier to the
 * mdast footnote definition.
 *
 * Footnotes come in two flavors:
 *
 * 1.  footnoteDefinition: a standalone footnote definition that is referenced
 *     by footnoteReference using the identifier. Looks like:
 *
 *         Some text[^1]
 *
 *         [^1]: A footnote definition.
 *
 * 2.  footnote: an inline footnote. We createDefault a footnote definition implicitly
 *     and compileNode them as a footnote reference. Looks like:
 *
 *         Some text [^ an inline footnote].
 *
 * For an interactive example, see:
 * https://astexplorer.net/#/gist/da878645e2b95030e1233407fd797f35/5e6eea3911b89f429f90330a8864820129eae1d5
 */
const addAllFnDefs = (p: PostAST): void => {
  let nextInlineId = 1;
  for (const { node } of unistNodes.preOrderGenerator(p.mdastNode)) {
    if (md.isFootnoteDefinition(node)) {
      p.addFootnoteDef(node);
    } else if (md.isFootnote(node)) {
      p.addInlineFootnoteDef(nextInlineId, node);
      nextInlineId++;
    }
  }
};
