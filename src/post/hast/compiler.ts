import { checkDefined } from '//asserts';
import { PostAST } from '//post/ast';
import { DocTemplate } from '//post/hast/doc_template';
import * as unist from 'unist';
import hastToHtml from 'hast-util-to-html';
import * as h from '//post/hast/nodes';

/**
 * Compiles a hast node into HTML.
 */
export class HastCompiler {
  private constructor() {}

  static create(): HastCompiler {
    return new HastCompiler();
  }

  /** Compiles node into a UTF-8 string. */
  compile(node: unist.Node, ast: PostAST): string {
    const pt = ast.metadata.postType;
    const template = checkDefined(
      DocTemplate.templates().get(pt),
      `No template found for post type: ${pt}`
    );
    const body = h.isRoot(node) ? node.children : [node];
    const doc = template.render(body);
    return hastToHtml(doc);
  }
}
