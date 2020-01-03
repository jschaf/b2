import * as unist from 'unist';
import hastToHtml from 'hast-util-to-html';
/**
 * Compiles a hast node into HTML.
 */
export class HastCompiler {
  private constructor() {}

  static create(): HastCompiler {
    return new HastCompiler();
  }

  /** Compiles node into a UTF-8 string as a buffer. */
  compile(node: unist.Node): string {
    // TODO: Use our own HTML serialization.
    return hastToHtml(node);
  }
}
