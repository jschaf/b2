declare module 'remark-rehype' {
  import { Plugin } from 'unified';

  interface RemarkRehype extends Plugin<[Partial<RemarkRehypeOptions>?]> {}

  type RemarkRehypeOptions = {};

  const remarkRehype: RemarkRehype;
  export = remarkRehype;
}
