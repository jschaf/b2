declare module '@starptech/prettyhtml-hast-to-html' {
  import * as unist from 'unist';

  interface PrettyHtmlOptions {
    singleQuote: boolean;
    printWidth: number;
    useTabs: boolean;
    tabWidth: number;
    wrapAttributes: boolean;
  }
  const toHtml: (
    node: unist.Node,
    options?: Partial<PrettyHtmlOptions>
  ) => string;
  export = toHtml;
}
