import { checkDefined, checkState } from '//asserts';
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

  private constructor(
    readonly mdastNode: unist.Node,
    readonly fnDefsById: Map<string, mdast.FootnoteDefinition>
  ) {}

  static create(md: unist.Node): PostAST {
    const fnDefs = extractFnDefs(md);
    // TODO: use actual vfile.
    return new PostAST(md, fnDefs);
  }

  static inlineFootnotePrefix = 'gen-';

  static newInlineFootnoteId(id: number): string {
    return PostAST.inlineFootnotePrefix + id;
  }
}

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
const extractFnDefs = (
  tree: unist.Node
): Map<string, mdast.FootnoteDefinition> => {
  const fnDefsById: Map<string, mdast.FootnoteDefinition> = new Map();
  const prefix = PostAST.inlineFootnotePrefix;

  // Extract footnote definitions first because the identifiers are defined in
  // the text so they take precedence over inline footnote definitions.
  for (const { node } of unistNodes.preOrderGenerator(tree)) {
    if (!md.isFootnoteDefinition(node)) {
      continue;
    }
    let id = node.identifier;
    checkDefined(id, 'Footnote definition must have an identifier');
    checkState(id !== '', 'Footnote definition identifier must not be empty');
    checkState(
      !id.startsWith(prefix),
      `Footnote definition must not start with ${prefix}`
    );
    if (fnDefsById.has(id)) {
      throw new Error(`Duplicate footnote definition identifier: '${id}'`);
    }
    fnDefsById.set(id, node);
  }

  // Extract inline footnotes.
  let nextInlineId = 1;
  for (const { node } of unistNodes.preOrderGenerator(tree)) {
    if (!md.isFootnote(node)) {
      continue;
    }
    const fnId = prefix + nextInlineId;
    // Store the ID so we can render the footnote reference.
    unistNodes.ensureDataAttr(node).data[
      PostAST.INLINE_FOOTNOTE_DATA_KEY
    ] = fnId;
    nextInlineId++;
    const fnDef = md.footnoteDef(fnId, [md.paragraph(node.children)]);
    fnDefsById.set(fnId, fnDef);
  }

  return fnDefsById;
};
