import { checkDefined } from '//asserts';
import { PostAST } from '//post/ast';
import { DocTemplate } from '//post/hast/doc_template';
import { StringBuilder } from '//strings';
import * as unist from 'unist';
import * as h from '//post/hast/nodes';
import * as nw from '//post/hast/node_writer';

export type NewNodeWriter = (c: HastWriter) => nw.HastNodeWriter;
export type NodeWriterEntries = [string, NewNodeWriter][];

export const newDefaultWriters: () => NodeWriterEntries = () => [
  ['doctype', nw.DoctypeWriter.create],
  ['element', nw.ElementWriter.create],
  ['raw', nw.RawWriter.create],
  ['root', nw.RootWriter.create],
  ['text', nw.TextWriter.create],
];

export class WriterContext {
  readonly indentLength: number = 2;
  private constructor(readonly postAST: PostAST, public indentLevel: number) {}

  static create(postAST: PostAST): WriterContext {
    return new WriterContext(postAST, 0);
  }

  incrementIndent(): void {
    this.indentLevel += 1;
  }

  decrementIndent(): void {
    this.indentLevel -= 1;
  }

  clone(): WriterContext {
    return new WriterContext(this.postAST, this.indentLevel);
  }
}

/**
 * HastWriter compiles a hast node into an HTML string.
 */
export class HastWriter {
  private readonly subWriters: Map<string, nw.HastNodeWriter> = new Map();

  private constructor(
    private readonly subWriterFactory: Map<string, NewNodeWriter>
  ) {}

  static create(writers: NodeWriterEntries): HastWriter {
    return new HastWriter(new Map<string, NewNodeWriter>(writers));
  }

  static createDefault(): HastWriter {
    return HastWriter.create(newDefaultWriters());
  }

  /** Compiles a hast node into a UTF-8 string. */
  write(node: unist.Node, ast: PostAST): string {
    const pt = ast.metadata.postType;
    const template = checkDefined(
      DocTemplate.templates().get(pt),
      `No template found for post type: ${pt}`
    );
    const body = h.isRoot(node) ? node.children : [node];
    const doc = template.render(body);
    const sb = StringBuilder.create();
    const ctx = WriterContext.create(ast);
    this.writeNode(doc, ctx, sb);
    return sb.toString();
  }

  writeNode(node: unist.Node, ctx: WriterContext, sb: StringBuilder): void {
    const w = this.getNodeWriter(node.type);
    w.writeNode(node, ctx.clone(), sb);
  }

  private getNodeWriter(type: string): nw.HastNodeWriter {
    // Check cache first.
    const w = this.subWriters.get(type);
    if (w !== undefined) {
      return w;
    }

    const newWriterFactory = checkDefined(
      this.subWriterFactory.get(type),
      `No hast compiler found for type: ${type}`
    );
    const newWriter = newWriterFactory(this);
    this.subWriters.set(type, newWriter);
    return newWriter;
  }
}
