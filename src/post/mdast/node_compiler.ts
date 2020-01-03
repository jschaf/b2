import { PostAST } from '//post/post_ast';
import * as unist from 'unist';

export interface MdastNodeCompiler {
  compileNode(node: unist.Node, ast: PostAST): unist.Node;
}
