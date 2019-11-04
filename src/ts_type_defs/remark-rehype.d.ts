declare module 'remark-rehype' {
  import {Plugin} from 'unified'

  interface RemarkRehype extends Plugin<[Partial<RemarkRehypeOptions>?]> {
  }

  interface RemarkRehypeOptions {}

  const remarkRehype: RemarkRehype;
  export = remarkRehype;
}
