import { hastElemWithProps, hastText } from '//post/mdast/hast_nodes';
import { checkNodeType } from '//post/mdast/md_nodes';
import { MdastNodeCompiler } from '//post/mdast/node_compiler';
import { PostAST } from '//post/post_ast';
import * as unist from 'unist';
import * as md_nodes from '//post/mdast/md_nodes';

export class FootnoteReferenceCompiler implements MdastNodeCompiler {
  private constructor() {}

  static create(): FootnoteReferenceCompiler {
    return new FootnoteReferenceCompiler();
  }

  compileNode(node: unist.Node, _postAST: PostAST): unist.Node {
    checkNodeType(node, 'footnoteReference', md_nodes.isFootnoteReference);
    const fnId = node.identifier;
    // There's also node.label which mdast defines as the original value of
    // the normalized node.identifier field.  We'll only use node.identifier
    // since I'm not sure when the label would ever be different than the ID.
    return FootnoteReferenceCompiler.makeHastNode(fnId);
  }

  static makeHastNode(fnId: string) {
    return hastElemWithProps('sup', { id: `fn-ref-${fnId}` }, [
      hastElemWithProps('a', { href: `#fn-${fnId}`, className: ['fn-ref'] }, [
        hastText(fnId),
      ]),
    ]);
  }
}
