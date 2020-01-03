import { checkDefined, checkState } from '//asserts';
import { MdastCompiler } from '//post/mdast/compiler';
import { checkNodeType } from '//post/mdast/md_nodes';
import { MdastNodeCompiler } from '//post/mdast/node_compiler';
import { PostAST } from '//post/post_ast';
import { mdFootnoteRef } from '//post/testing/markdown_nodes';
import { isString } from '//strings';
import * as unist from 'unist';
import * as md_nodes from '//post/mdast/md_nodes';

/**
 * An inline footnote definition.
 * https://github.com/syntax-tree/mdast#footnote
 */
export class FootnoteCompiler implements MdastNodeCompiler {
  private constructor(private readonly compiler: MdastCompiler) {}

  static create(compiler: MdastCompiler): FootnoteCompiler {
    return new FootnoteCompiler(compiler);
  }

  compileNode(node: unist.Node, postAST: PostAST): unist.Node {
    checkNodeType(node, 'footnote', md_nodes.isFootnote);
    const data = checkDefined(
      node.data,
      'Expected data attr to exist on footnote type.'
    );
    let key = PostAST.INLINE_FOOTNOTE_DATA_KEY;
    const fnId = checkDefined(
      data[key],
      `Expected a footnote ID to exist on data attr with key: ${key}`
    );
    checkState(
      isString(fnId),
      `Expected data.${key} to be a string but was ${fnId}`
    );
    const fnRef = mdFootnoteRef(fnId);
    return this.compiler.compileNode(fnRef, postAST);
  }
}
