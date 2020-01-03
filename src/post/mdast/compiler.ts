import { checkArg, checkDefined } from '//asserts';
import { PostAST } from '//post/post_ast';
import {
  BlockquoteCompiler,
  BreakCompiler,
  CodeCompiler,
  DeleteCompiler,
  EmphasisCompiler,
  FootnoteCompiler,
  FootnoteReferenceCompiler,
  HeadingCompiler, InlineCodeCompiler,
  MdastNodeCompiler,
  ParagraphCompiler,
  RootCompiler,
  TextCompiler,
  TomlCompiler,
} from '//post/mdast/node_compiler';
import * as unist from 'unist';
import * as unistNodes from '//unist/nodes';

export type NewNodeCompiler = (parent: MdastCompiler) => MdastNodeCompiler;
export type NodeCompilerEntries = [string, NewNodeCompiler][];

export const newDefaultCompilers: () => NodeCompilerEntries = () => [
  ['blockquote', BlockquoteCompiler.create],
  ['break', BreakCompiler.create],
  ['code', CodeCompiler.create],
  ['delete', DeleteCompiler.create],
  ['emphasis', EmphasisCompiler.create],
  ['footnote', FootnoteCompiler.create],
  ['footnoteReference', FootnoteReferenceCompiler.create],
  ['heading', HeadingCompiler.create],
  ['inlineCode', InlineCodeCompiler.create],
  ['paragraph', ParagraphCompiler.create],
  ['text', TextCompiler.create],
  ['toml', TomlCompiler.create],
  ['root', RootCompiler.create],
];

/** Compiles an mdast node into a hast node. */
export class MdastCompiler {
  private readonly subCompilers: Map<string, MdastNodeCompiler> = new Map();

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

  private getNodeCompiler(type: string): MdastNodeCompiler {
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
