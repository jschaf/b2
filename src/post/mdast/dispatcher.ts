import {
  BreakCompiler,
  CodeCompiler,
  DeleteCompiler,
  EmphasisCompiler,
  FootnoteCompiler,
  FootnoteReferenceCompiler,
  HeadingCompiler,
  MdastNodeCompiler,
  ParagraphCompiler,
  RootCompiler,
  TextCompiler,
} from '//post/mdast/node_compiler';
import * as unist from 'unist';

/** Chooses the correct compiler for an mdast node. */
export class MdastDispatcher {
  private static instance: MdastDispatcher | null = null;

  private readonly renderersByType: Map<string, MdastNodeCompiler>;

  private constructor(renderers: Map<string, MdastNodeCompiler>) {
    this.renderersByType = renderers;
  }

  static defaultInstance(): MdastDispatcher {
    if (MdastDispatcher.instance === null) {
      MdastDispatcher.instance = new MdastDispatcher(new Map());
      const rd = MdastDispatcher.instance;
      const defaults: [string, MdastNodeCompiler][] = [
        ['break', BreakCompiler.create()],
        ['code', CodeCompiler.create()],
        ['delete', DeleteCompiler.create(rd)],
        ['emphasis', EmphasisCompiler.create(rd)],
        ['footnote', FootnoteCompiler.create(rd)],
        ['footnoteReference', FootnoteReferenceCompiler.create()],
        ['heading', HeadingCompiler.create(rd)],
        ['paragraph', ParagraphCompiler.create(rd)],
        ['text', TextCompiler.create()],
        ['root', RootCompiler.create(rd)],
      ];
      for (const [type, r] of defaults) {
        rd.renderersByType.set(type, r);
      }
    }
    return MdastDispatcher.instance;
  }

  chooseCompiler(node: unist.Node): MdastNodeCompiler | undefined {
    const r = this.renderersByType.get(node.type);
    const name = r === undefined ? '<undefined>' : r.constructor.name;
    console.log(`!!! Dispatching node: ${node.type} to ${name}`);
    return r;
  }
}
