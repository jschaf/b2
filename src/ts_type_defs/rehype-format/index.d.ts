declare module 'rehype-format' {
  import { Plugin } from 'unified';

  interface RehypeFormat extends Plugin<[Partial<RehypeFormatOptions>?]> {}

  type RehypeFormatOptions = {
    indent: number | string;
    indentInitial: string;
    blanks: string[];
  };

  const rehypeFormat: RehypeFormat;
  export = rehypeFormat;
}
