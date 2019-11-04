declare module 'rehype-parse' {
  import { Plugin } from 'unified';

  interface RehypeParse extends Plugin<[Partial<RehypeParseOptions>?]> {}

  interface RehypeParseOptions {}

  const rehypeParse: RehypeParse;
  export = rehypeParse;
}
