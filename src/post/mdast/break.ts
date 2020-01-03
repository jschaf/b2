import { hastElem } from '//post/mdast/hast_nodes';
import { checkNodeType } from '//post/mdast/md_nodes';
import { PostAST } from '//post/post_ast';
import { MdastNodeCompiler } from '//post/mdast/node_compiler';
import * as unist from 'unist';
import * as md_nodes from '//post/mdast/md_nodes';

export class BreakCompiler implements MdastNodeCompiler {
  private constructor() {}

  static create(): BreakCompiler {
    return new BreakCompiler();
  }

  compileNode(node: unist.Node, _postAST: PostAST): unist.Node {
    checkNodeType(node, 'break', md_nodes.isBreak);
    return hastElem('break');
  }
}
