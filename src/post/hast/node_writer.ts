import { PostAST } from '//post/ast';
import { HastCompiler } from '//post/hast/compiler';
import { isLiteralElem, isParentElem } from '//post/hast/nodes';
import { StringBuilder } from '//strings';
import * as unist from 'unist';
import * as h from '//post/hast/nodes';
import * as un from '//unist/nodes';

/** Compiler for a single mdast node. */
export interface HastNodeWriter {
  writeNode(node: unist.Node, ast: PostAST, sb: StringBuilder): void;
}

/**
 * Compiles an hast doctype node to an HTML string, like:
 *
 *     <!doctype html>
 *
 * https://github.com/syntax-tree/hast#doctype
 */
export class DoctypeWriter implements HastNodeWriter {
  private constructor() {}

  static create(): DoctypeWriter {
    return new DoctypeWriter();
  }

  writeNode(node: unist.Node, _postAST: PostAST, sb: StringBuilder): void {
    h.checkType(node, 'doctype', h.isDoctype);
    sb.writeString('<!doctype html>\n');
  }
}

/**
 * Compiles an hast comment node to an HTML string.
 *
 * https://github.com/syntax-tree/hast#comment
 */
export class CommentWriter implements HastNodeWriter {
  private constructor() {}

  static create(): CommentWriter {
    return new CommentWriter();
  }

  writeNode(node: unist.Node, _postAST: PostAST, sb: StringBuilder): void {
    h.checkType(node, 'comment', h.isComment);
    sb.writeString(`<!-- ${node.value} -->`);
  }
}

/**
 * Compiles an hast element node to an HTML string.
 *
 * https://github.com/syntax-tree/hast#element
 */
export class ElementWriter implements HastNodeWriter {
  private constructor(private readonly compiler: HastCompiler) {}

  static create(hc: HastCompiler): ElementWriter {
    return new ElementWriter(hc);
  }

  writeNode(node: unist.Node, ast: PostAST, sb: StringBuilder): void {
    h.checkType(node, 'element', h.isElem);
    sb.writeString(`<${node.tagName}>`);

    // TODO: write attributes.
    if (isParentElem(node)) {
      for (const child of node.children) {
        this.compiler.writeNode(child, ast, sb);
      }
    }

    if (isLiteralElem(node)) {
      // TODO: Escape everything except style and script tags.
      sb.writeString(node.value);
    }

    sb.writeString(`</${node.tagName}>`);
  }
}

/** Compiles an hast raw node to an HTML string. */
export class RawWriter implements HastNodeWriter {
  private constructor() {}

  static create(): RawWriter {
    return new RawWriter();
  }

  writeNode(node: unist.Node, _postAST: PostAST, sb: StringBuilder): void {
    h.checkType(node, 'raw', h.isRaw);
    sb.writeString(node.value + '\n');
  }
}

/**
 * Compiles an hast root node to an HTML string.
 *
 * https://github.com/syntax-tree/hast#root
 */
export class RootWriter implements HastNodeWriter {
  private constructor(private readonly compiler: HastCompiler) {}

  static create(hc: HastCompiler): RootWriter {
    return new RootWriter(hc);
  }

  writeNode(node: unist.Node, ast: PostAST, sb: StringBuilder): void {
    h.checkType(node, 'root', h.isRoot);
    for (const child of node.children) {
      this.compiler.writeNode(child, ast, sb);
    }
  }
}

/**
 * Compiles an hast text node to an HTML string.
 *
 * https://github.com/syntax-tree/hast#text
 */
export class TextWriter implements HastNodeWriter {
  private constructor() {}

  static create(): TextWriter {
    return new TextWriter();
  }

  writeNode(node: unist.Node, _postAST: PostAST, sb: StringBuilder): void {
    h.checkType(node, 'text', un.isText);
    sb.writeString(node.value);
  }
}
