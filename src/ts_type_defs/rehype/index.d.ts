declare module 'rehype' {
  import * as unified from 'unified';

  interface Rehype extends unified.Plugin<[Partial<RehypeOptions>?]> {}

  type RehypeOptions = {};

  function rehype<P = unified.Settings>(): unified.Processor<P>;
  export = rehype;
}
