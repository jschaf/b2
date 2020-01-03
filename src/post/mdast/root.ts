import { MdastCompiler } from '//post/mdast/compiler';
import { hastElem } from '//post/mdast/hast_nodes';
import { PostAST } from '//post/post_ast';
import { MdastNodeCompiler } from '//post/mdast/node_compiler';
import * as unist from 'unist';
import * as mdast from 'mdast';

export class RootCompiler implements MdastNodeCompiler {
  private constructor(private readonly compiler: MdastCompiler) {}

  static create(compiler: MdastCompiler): RootCompiler {
    return new RootCompiler(compiler);
  }

  compileNode(node: mdast.Root, postAST: PostAST): unist.Node {
    const children = this.compiler.compileChildren(node, postAST);
    return hastElem('body', children);
  }
}
