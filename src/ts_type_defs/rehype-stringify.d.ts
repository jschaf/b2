declare module 'rehype-stringify' {
  import { Plugin } from 'unified';

  interface RehypeStringify
    extends Plugin<[Partial<RehypeStringifyOptions>?]> {}

  interface RehypeStringifyOptions {}

  const rehypeStringify: RehypeStringify;
  export = rehypeStringify;
}
