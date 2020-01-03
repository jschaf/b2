import { hastElem, hastElemWithProps, hastText } from '//post/mdast/hast_nodes';
import { checkNodeType } from '//post/mdast/md_nodes';
import * as md_nodes from '//post/mdast/md_nodes';
import { PostAST } from '//post/post_ast';
import { MdastNodeCompiler } from '//post/mdast/node_compiler';
import { isString } from '//strings';
import * as unist from 'unist';

export class CodeCompiler implements MdastNodeCompiler {
  private constructor() {}

  static create(): CodeCompiler {
    return new CodeCompiler();
  }

  compileNode(node: unist.Node, _postAST: PostAST): unist.Node {
    checkNodeType(node, 'code', md_nodes.isCode);
    const props: Record<string, unknown> = {};
    if (isString(node.lang) && node.lang !== '') {
      props.className = ['lang-' + node.lang.trim()];
    }
    return hastElem('pre', [
      hastElemWithProps('code', props, [hastText(node.value)]),
    ]);
  }
}
