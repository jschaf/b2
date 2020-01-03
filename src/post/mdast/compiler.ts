import { checkArg, checkDefined } from '//asserts';
import { PostAST } from '//post/post_ast';
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
  ['link', nc.LinkCompiler.create],
  ['paragraph', nc.ParagraphCompiler.create],
  ['strong', nc.StrongCompiler.create],
  ['text', nc.TextCompiler.create],
  ['toml', nc.TomlCompiler.create],
  ['root', nc.RootCompiler.create],
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

  /** Compiles the entire post AST into a single hast node. */
  compile(postAST: PostAST): unist.Node {
    checkArg(postAST.mdastNode.type !== unistNodes.IGNORED_TYPE);
    return this.compileNode(postAST.mdastNode, postAST);
  }

  /** Compiles a single mdast node from a  post AST into a single hast node. */
  compileNode(node: unist.Node, postAST: PostAST): unist.Node {
    const c = this.getNodeCompiler(node.type);
    return c.compileNode(node, postAST);
  }

  /** Compiles all children of an mdast node into an array of hast nodes. */
  compileChildren(parent: unist.Parent, postAST: PostAST): unist.Node[] {
    const results: unist.Node[] = [];
    for (const child of parent.children) {
      const r = this.compileNode(child, postAST);
      if (r.type === unistNodes.IGNORED_TYPE) {
        continue;
      }
      results.push(r);
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
