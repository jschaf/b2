import * as objects from '//objects';
import { AttrWriter } from '//post/hast/attr_writer';
import * as h from '//post/hast/nodes';
import { isParentElem } from '//post/hast/nodes';
import { HastWriter, WriterContext } from '//post/hast/writer';
import { HTMLEscaper } from '//post/html/escapers';
import { StringBuilder } from '//strings';
import { isParent } from '//unist/nodes';
import * as un from '//unist/nodes';
import * as unist from 'unist';

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

    if (isBlockTag(node)) {
      sb.writeString('\n');

      // Head and body aren't indented.
      if (node.tagName !== 'head' && node.tagName !== 'body') {
        ctx.incrementIndent();
        sb.writeString(newIndentString(ctx));
      }
    }

    sb.writeString(`<${node.tagName}`);
    const p = node.properties;
    if (objects.isObject(p) && !objects.isEmpty(p)) {
      sb.writeString(' ');
      this.attrWriter.writeElemProps(p, sb);
    }
    sb.writeString('>');

    if (isVoidTag(node)) {
      return;
    } else if (isParentElem(node)) {
      for (const child of node.children) {
        this.compiler.writeNode(child, ctx, sb);
      }
    } else {
      throw new Error(`unknown element: ${node.tagName}`);
    }

    if (isAnyChildBlockTag(node)) {
      sb.writeString('\n');
      sb.writeString(newIndentString(ctx));
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

  writeNode(node: unist.Node, ctx: WriterContext, sb: StringBuilder): void {
    h.checkType(node, 'text', un.isText);
    const parent = ctx.ancestors[ctx.ancestors.length - 1];
    const parentTag = parent === undefined ? '<unknown>' : parent.tagName;

    switch (parentTag) {
      case 'script':
      case 'style':
        sb.writeString(node.value);
        break;

      default:
        const e = HTMLEscaper.escape(node.value);
        sb.writeString(e);
        break;
    }
  }
}

const blockTags = [
  'article',
  'blockquote',
  'body',
  'div',
  'head',
  'heading',
  'header',
  'h1',
  'h2',
  'h3',
  'h4',
  'h5',
  'h6',
  'footer',
  'li',
  'link',
  'main',
  'meta',
  'ol',
  'p',
  'pre',
  'section',
  'script',
  'ul',
];

const isBlockTag = (n: unist.Node): boolean => {
  return h.isElem(n) && blockTags.includes(n.tagName);
};

/** Returns true if any child of the node is a block tag. */
const isAnyChildBlockTag = (n: unist.Node): boolean => {
  if (isParent(n)) {
    for (const child of n.children) {
      if (isBlockTag(child)) {
        return true;
      }
    }
  }
  return false;
};

const voidTags: readonly string[] = <const>[
  'area',
  'base',
  'br',
  'col',
  'embed',
  'hr',
  'img',
  'input',
  'link',
  'meta',
  'param',
  'source',
  'track',
  'wbr',
];

/**
 *
 * https://html.spec.whatwg.org/multipage/syntax.html#start-tags
 */
const isVoidTag = (n: unist.Node): boolean => {
  return h.isElem(n) && voidTags.includes(n.tagName);
};

const newIndentString = (c: WriterContext): string => {
  const l = c.indentLevel * c.indentLength;
  return ' '.repeat(l);
};
