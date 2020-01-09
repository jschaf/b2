import { checkDefined } from '//asserts';
import { PostAST } from '//post/ast';
import { DocTemplate } from '//post/hast/doc_template';
import { StringBuilder } from '//strings';
import unified from 'unified';
import * as unist from 'unist';
import rehype from 'rehype';
import rehypeFormat from 'rehype-format';
import * as h from '//post/hast/nodes';
import * as nw from '//post/hast/node_writer';

export type NewNodeWriter = (c: HastCompiler) => nw.HastNodeWriter;
export type NodeWriterEntries = [string, NewNodeWriter][];

export const newDefaultWriters: () => NodeWriterEntries = () => [
  ['doctype', nw.DoctypeWriter.create],
  ['raw', nw.RawWriter.create],
  ['root', nw.RootWriter.create],
  ['text', nw.TextWriter.create],
];

/**
 * Compiles a hast node into HTML.
 */
export class HastCompiler {
  private readonly processor: unified.Processor;
  private readonly subWriters: Map<string, nw.HastNodeWriter> = new Map();

  private constructor(
    private readonly subWriterFactory: Map<string, NewNodeWriter>
  ) {
    this.processor = rehype().use(rehypeFormat);
  }

  static create(writers: NodeWriterEntries): HastCompiler {
    return new HastCompiler(new Map<string, NewNodeWriter>(writers));
  }

  static createDefault(): HastCompiler {
    return HastCompiler.create(newDefaultWriters());
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
    return this.processor.stringify(doc);
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
