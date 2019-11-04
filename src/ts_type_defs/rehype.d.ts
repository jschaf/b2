declare module 'rehype' {
  import { Plugin } from 'unified';

  interface Rehype extends Plugin<[Partial<RehypeOptions>?]> {}

  interface RehypeOptions {}

  const rehype: Rehype;
  export = rehype;
}
