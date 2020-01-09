import { PostAST } from '//post/ast';
import { StringBuilder } from '//strings';
import * as unist from 'unist';
import * as h from '//post/hast/nodes';
import * as un from '//unist/nodes';

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

/**
 * Compiles an hast raw node to an HTML string.
 *
 * https://github.com/syntax-tree/hast#raw
 */
export class RawWriter implements HastNodeWriter {
  private constructor(private readonly sb: StringBuilder) {}

  static create(sb: StringBuilder): RawWriter {
    return new RawWriter(sb);
  }

  writeNode(node: unist.Node, _postAST: PostAST): void {
    h.checkType(node, 'raw', h.isRaw);
    this.sb.writeString(node.value + '\n');
  }
}

/**
 * Compiles an hast text node to an HTML string.
 *
 * https://github.com/syntax-tree/hast#text
 */
export class TextWriter implements HastNodeWriter {
  private constructor(private readonly sb: StringBuilder) {}

  static create(sb: StringBuilder): TextWriter {
    return new TextWriter(sb);
  }

  writeNode(node: unist.Node, _postAST: PostAST): void {
    h.checkType(node, 'text', un.isText);
    this.sb.writeString(node.value);
  }
}
