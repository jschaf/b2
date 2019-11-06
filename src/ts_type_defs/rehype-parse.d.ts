declare module 'rehype-parse' {
  import { Plugin } from 'unified';

  interface RehypeParse extends Plugin<[Partial<RehypeParseOptions>?]> {}

  type RehypeParseOptions = {};

  const rehypeParse: RehypeParse;
  export = rehypeParse;
}
