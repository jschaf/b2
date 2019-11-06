declare module 'remark-stringify' {
  import { Plugin } from 'unified';

  interface RemarkStringify
    extends Plugin<[Partial<RemarkStringifyOptions>?]> {}

  type RemarkStringifyOptions = {};

  const remarkStringify: RemarkStringify;
  export = remarkStringify;
}
