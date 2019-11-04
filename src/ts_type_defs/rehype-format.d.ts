declare module 'rehype-format' {
  import { Plugin } from 'unified';

  interface RehypeFormat extends Plugin<[Partial<RehypeFormatOptions>?]> {}

  interface RehypeFormatOptions {}

  const rehypeFormat: RehypeFormat;
  export = rehypeFormat;
}
