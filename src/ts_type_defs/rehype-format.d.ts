declare module 'rehype-format' {
  import { Plugin } from 'unified';

  interface RehypeFormat extends Plugin<[Partial<RehypeFormatOptions>?]> {}

  type RehypeFormatOptions = {};

  const rehypeFormat: RehypeFormat;
  export = rehypeFormat;
}
