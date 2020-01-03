declare module 'hast-util-to-html' {
  import * as unist from 'unist';

  interface StringifyEntitiesOptions {
    escapeOnly?: boolean;
    subset?: string[];
    useNamedReferences?: boolean;
    useShortestReferences?: boolean;
    omitOptionalSemicolons?: boolean;
    attribute?: boolean;
  }

  interface Options {
    space?: boolean;
    entities?: StringifyEntitiesOptions;
    voids?: string[];
    upperDoctype?: boolean;
    quote?: '"' | "'";
    quoteSmart?: boolean;
    preferUnquoted?: boolean;
    omitOptionalTags?: boolean;
    collapseEmptyAttributes?: boolean;
    closeSelfClosing?: boolean;
    tightSelfClosing?: boolean;
    tightCommaSeparatedLists?: boolean;
    tightAttributes?: boolean;
    tightDoctype?: boolean;
    allowParseErrors?: boolean;
    allowDangerousCharacters?: boolean;
    allowDangerousHTML?: boolean;
  }

  const toHtml: <T extends unist.Node>(
    tree: unist.Node,
    options?: Options
  ) => string;
  export = toHtml;
}
