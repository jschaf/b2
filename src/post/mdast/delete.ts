import { MdastCompiler } from '//post/mdast/compiler';
import { hastElem } from '//post/mdast/hast_nodes';
import { checkNodeType } from '//post/mdast/md_nodes';
import { PostAST } from '//post/post_ast';
import { MdastNodeCompiler } from '//post/mdast/node_compiler';
import * as unist from 'unist';
import * as md_nodes from '//post/mdast/md_nodes';

export class DeleteCompiler implements MdastNodeCompiler {
  private constructor(private readonly compiler: MdastCompiler) {}

  static create(compiler: MdastCompiler): DeleteCompiler {
    return new DeleteCompiler(compiler);
  }

  compileNode(node: unist.Node, postAST: PostAST): unist.Node {
    checkNodeType(node, 'delete', md_nodes.isDelete);
    const children = this.compiler.compileChildren(node, postAST);
    return hastElem('del', children);
  }
}
