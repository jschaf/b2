declare module 'rehype-katex' {
  import {Plugin} from 'unified'

  interface RehypeKatex extends Plugin<[Partial<RehypeKatexOptions>?]> {
  }

  interface RehypeKatexOptions {}

  const rehypeKatex: RehypeKatex;
  export = rehypeKatex;
}
