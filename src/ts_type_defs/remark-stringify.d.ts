declare module 'remark-stringify' {
  import { Plugin } from 'unified';

  interface RemarkStringify
    extends Plugin<[Partial<RemarkStringifyOptions>?]> {}

  interface RemarkStringifyOptions {}

  const remarkStringify: RemarkStringify;
  export = remarkStringify;
}
