import { PostAST } from '//post/ast';
import { StringBuilder } from '//strings';
import * as unist from 'unist';
import * as h from '//post/hast/nodes';

/** Compiler for a single mdast node. */
export interface HastNodeWriter {
  writeNode(node: unist.Node, ast: PostAST): void;
}

/**
 * Compiles an hast doctype node to an HTML string, like:
 *
 *     <!doctype html>
 *
 * https://github.com/syntax-tree/hast#doctype
 */
export class DoctypeWriter implements HastNodeWriter {
  private constructor(private readonly sb: StringBuilder) {}

  static create(sb: StringBuilder): DoctypeWriter {
    return new DoctypeWriter(sb);
  }

  writeNode(node: unist.Node, _postAST: PostAST): void {
    h.checkType(node, 'doctype', h.isDoctype);
    this.sb.writeString('<!doctype html>\n');
  }
}
