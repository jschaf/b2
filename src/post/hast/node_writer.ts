import { AttrWriter } from '//post/hast/attr_writer';
import { HastWriter, WriterContext } from '//post/hast/writer';
import { isLiteralElem, isParentElem } from '//post/hast/nodes';
import { StringBuilder } from '//strings';
import * as unist from 'unist';
import * as h from '//post/hast/nodes';
import * as un from '//unist/nodes';
import * as objects from '//objects';

/** Compiler for a single mdast node. */
export interface HastNodeWriter {
  writeNode(node: unist.Node, ctx: WriterContext, sb: StringBuilder): void;
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

  writeNode(node: unist.Node, _ctx: WriterContext, sb: StringBuilder): void {
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

  writeNode(node: unist.Node, _ctx: WriterContext, sb: StringBuilder): void {
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
  private attrWriter = AttrWriter.create();
  private constructor(private readonly compiler: HastWriter) {}

  static create(hc: HastWriter): ElementWriter {
    return new ElementWriter(hc);
  }

  writeNode(node: unist.Node, ctx: WriterContext, sb: StringBuilder): void {
    h.checkType(node, 'element', h.isElem);
    sb.writeString(`<${node.tagName}`);
    const p = node.properties;
    if (objects.isObject(p) && !objects.isEmpty(p)) {
      sb.writeString(' ');
      this.attrWriter.writeElemProps(p, sb);
    }
    sb.writeString('>');

    if (isParentElem(node)) {
      for (const child of node.children) {
        this.compiler.writeNode(child, ctx, sb);
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

  writeNode(node: unist.Node, _ctx: WriterContext, sb: StringBuilder): void {
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
  private constructor(private readonly compiler: HastWriter) {}

  static create(hc: HastWriter): RootWriter {
    return new RootWriter(hc);
  }

  writeNode(node: unist.Node, ctx: WriterContext, sb: StringBuilder): void {
    h.checkType(node, 'root', h.isRoot);
    for (const child of node.children) {
      this.compiler.writeNode(child, ctx, sb);
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

  writeNode(node: unist.Node, _ctx: WriterContext, sb: StringBuilder): void {
    h.checkType(node, 'text', un.isText);
    sb.writeString(node.value);
  }
}
