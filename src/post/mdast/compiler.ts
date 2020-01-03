import { checkDefined } from '//asserts';
import { BreakCompiler } from '//post/mdast/break';
import { CodeCompiler } from '//post/mdast/code';
import { DeleteCompiler } from '//post/mdast/delete';
import { EmphasisCompiler } from '//post/mdast/emphasis';
import { FootnoteCompiler } from '//post/mdast/footnote';
import { FootnoteReferenceCompiler } from '//post/mdast/footnote_reference';
import { HeadingCompiler } from '//post/mdast/heading';
import { ParagraphCompiler } from '//post/mdast/paragraph';
import { RootCompiler } from '//post/mdast/root';
import { TextCompiler } from '//post/mdast/text';
import { PostAST } from '//post/post_ast';
import { MdastNodeCompiler } from '//post/mdast/node_compiler';
import * as unist from 'unist';

type NewNodeCompiler = (parent: MdastCompiler) => MdastNodeCompiler;
type NodeCompilerMap = Map<string, NewNodeCompiler>;

export const newDefaultCompilers: () => Map<string, NewNodeCompiler> = () =>
  new Map<string, NewNodeCompiler>([
    ['break', () => BreakCompiler.create()],
    ['code', () => CodeCompiler.create()],
    ['delete', c => DeleteCompiler.create(c)],
    ['emphasis', c => EmphasisCompiler.create(c)],
    ['footnote', c => FootnoteCompiler.create(c)],
    ['footnoteReference', () => FootnoteReferenceCompiler.create()],
    ['heading', c => HeadingCompiler.create(c)],
    ['paragraph', c => ParagraphCompiler.create(c)],
    ['text', () => TextCompiler.create()],
    ['root', c => RootCompiler.create(c)],
  ]);

/** Compiles an mdast node into a hast node. */
export class MdastCompiler {
  private readonly subCompilers: Map<string, MdastNodeCompiler> = new Map();

  private constructor(private readonly subCompilersFactory: NodeCompilerMap) {}

  static createDefault(): MdastCompiler {
    return MdastCompiler.create(newDefaultCompilers());
  }

  static create(m: NodeCompilerMap): MdastCompiler {
    return new MdastCompiler(m);
  }

  compile(postAST: PostAST): unist.Node {
    return this.compileNode(postAST.mdastNode, postAST);
  }

  compileNode(node: unist.Node, postAST: PostAST): unist.Node {
    const c = this.getNodeCompiler(node.type);
    return c.compileNode(node, postAST);
  }

  compileChildren(parent: unist.Parent, postAST: PostAST): unist.Node[] {
    return parent.children.map(c => this.compileNode(c, postAST));
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
