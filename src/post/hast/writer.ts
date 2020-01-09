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

/**
 * Compiles a hast node into HTML.
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

  /** Compiles node into a UTF-8 string. */
  write(node: unist.Node, ast: PostAST): string {
    const pt = ast.metadata.postType;
    const template = checkDefined(
      DocTemplate.templates().get(pt),
      `No template found for post type: ${pt}`
    );
    const body = h.isRoot(node) ? node.children : [node];
    const doc = template.render(body);
    const sb = StringBuilder.create();
    this.writeNode(doc, ast, sb);
    return sb.toString();
  }

  writeNode(node: unist.Node, ast: PostAST, sb: StringBuilder): void {
    const w = this.getNodeWriter(node.type);
    w.writeNode(node, ast, sb);
  }

  private getNodeWriter(type: string): nw.HastNodeWriter {
    // Check cache first.
    const w = this.subWriters.get(type);
    if (w !== undefined) {
      return w;
    }

    const newWriter = checkDefined(
      this.subWriterFactory.get(type),
      `No hast compiler found for type: ${type}`
    )(this);
    this.subWriters.set(type, newWriter);
    return newWriter;
  }
}
