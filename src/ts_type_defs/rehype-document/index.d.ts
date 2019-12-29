declare module 'rehype-document' {
  import { Plugin } from 'unified';

  interface RemarkRehype extends Plugin<[Partial<RemarkRehypeOptions>?]> {}

  type RemarkRehypeOptions = {};

  const rehypeDocument: RemarkRehype;
  export = rehypeDocument;
}
