import { checkArg, checkDefined } from '//asserts';
import { PostAST } from '//post/ast';
import * as nc from '//post/mdast/node_compiler';
import * as unist from 'unist';
import * as unistNodes from '//unist/nodes';

export type NewNodeCompiler = (parent: MdastCompiler) => nc.MdastNodeCompiler;
export type NodeCompilerEntries = [string, NewNodeCompiler][];

export const newDefaultCompilers: () => NodeCompilerEntries = () => [
  ['blockquote', nc.BlockquoteCompiler.create],
  ['break', nc.BreakCompiler.create],
  ['code', nc.CodeCompiler.create],
  ['delete', nc.DeleteCompiler.create],
  ['emphasis', nc.EmphasisCompiler.create],
  ['footnote', nc.FootnoteCompiler.create],
  ['footnoteReference', nc.FootnoteReferenceCompiler.create],
  ['heading', nc.HeadingCompiler.create],
  ['html', nc.HTMLCompiler.create],
  ['image', nc.ImageCompiler.create],
  ['imageReference', nc.ImageReferenceCompiler.create],
  ['inlineCode', nc.InlineCodeCompiler.create],
  ['inlineMath', nc.InlineMathCompiler.create],
  ['link', nc.LinkCompiler.create],
  ['linkReference', nc.LinkReferenceCompiler.create],
  ['list', nc.ListCompiler.create],
  ['listItem', nc.ListItemCompiler.create],
  ['paragraph', nc.ParagraphCompiler.create],
  ['strong', nc.StrongCompiler.create],
  ['table', nc.TableCompiler.create],
  ['text', nc.TextCompiler.create],
  ['thematicBreak', nc.ThematicBreakCompiler.create],
  ['root', nc.RootCompiler.create],

  // Noops
  // definition is handled by postAST.
  ['definition', nc.NoopCompiler.create],
  // footnote definition is handled by postAST.
  ['footnoteDefinition', nc.NoopCompiler.create],
  ['toml', nc.NoopCompiler.create],
  ['yaml', nc.NoopCompiler.create],
];

/** Compiles an mdast node into a hast node. */
export class MdastCompiler {
  private readonly subCompilers: Map<string, nc.MdastNodeCompiler> = new Map();

  private constructor(
    private readonly subCompilersFactory: Map<string, NewNodeCompiler>
  ) {}

  static createDefault(): MdastCompiler {
    return MdastCompiler.create(newDefaultCompilers());
  }

  static create(m: [string, NewNodeCompiler][]): MdastCompiler {
    return new MdastCompiler(new Map<string, NewNodeCompiler>(m));
  }

  /** Compiles the entire post AST into an array of hast nodes. */
  compile(postAST: PostAST): unist.Node[] {
    checkArg(postAST.mdastNode.type !== unistNodes.IGNORED_TYPE);
    return this.compileNode(postAST.mdastNode, postAST);
  }

  /**
   * Compiles a single mdast node from a post AST into an array of hast nodes.
   */
  compileNode(node: unist.Node, postAST: PostAST): unist.Node[] {
    const c = this.getNodeCompiler(node.type);
    return c.compileNode(node, postAST);
  }

  /** Compiles all children of an mdast node into an array of hast nodes. */
  compileChildren(parent: unist.Parent, postAST: PostAST): unist.Node[] {
    const results: unist.Node[] = [];
    for (const child of parent.children) {
      const rs = this.compileNode(child, postAST);
      for (const r of rs) {
        results.push(r);
      }
    }
    return results;
  }

  private getNodeCompiler(type: string): nc.MdastNodeCompiler {
    const c = this.subCompilers.get(type);
    if (c !== undefined) {
      return c;
    }
    const newCompiler = checkDefined(
      this.subCompilersFactory.get(type),
      `No mdast compiler found for type: ${type}`
    )(this);
    this.subCompilers.set(type, newCompiler);
    return newCompiler;
  }
}
