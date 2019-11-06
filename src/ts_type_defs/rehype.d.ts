declare module 'rehype' {
  import { Plugin } from 'unified';

  interface Rehype extends Plugin<[Partial<RehypeOptions>?]> {}

  type RehypeOptions = {};

  const rehype: Rehype;
  export = rehype;
}
