import { BreakCompiler } from '//post/mdast/break';
import { CodeCompiler } from '//post/mdast/code';
import { DeleteCompiler } from '//post/mdast/delete';
import { EmphasisCompiler } from '//post/mdast/emphasis';
import { FootnoteCompiler } from '//post/mdast/footnote';
import { FootnoteReferenceCompiler } from '//post/mdast/footnote_reference';
import { HeadingCompiler } from '//post/mdast/heading';
import { ParagraphCompiler } from '//post/mdast/paragraph';
import { MdastNodeCompiler } from '//post/mdast/node_compiler';
import { RootCompiler } from '//post/mdast/root';
import { TextCompiler } from '//post/mdast/text';
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
