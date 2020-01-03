import { MdastCompiler } from '//post/mdast/compiler';
import { hastElem } from '//post/mdast/hast_nodes';
import { checkNodeType } from '//post/mdast/md_nodes';
import { PostAST } from '//post/post_ast';
import { MdastNodeCompiler } from '//post/mdast/node_compiler';
import * as unist from 'unist';
import * as md_nodes from '//post/mdast/md_nodes';

export class BlockquoteCompiler implements MdastNodeCompiler {
  private constructor(private readonly compiler: MdastCompiler) {}

  static create(compiler: MdastCompiler): BlockquoteCompiler {
    return new BlockquoteCompiler(compiler);
  }

  compileNode(node: unist.Node, postAST: PostAST): unist.Node {
    checkNodeType(node, 'blockquote', md_nodes.isBlockquote);
    const children = this.compiler.compileChildren(node, postAST);
    return hastElem('blockquote', children);
  }
}
