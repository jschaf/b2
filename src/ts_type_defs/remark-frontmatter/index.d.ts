declare module 'remark-frontmatter' {
  import { Plugin } from 'unified';

  interface RemarkFrontmatter extends Plugin<Array<RemarkFrontmatterOptions>> {}

  type RemarkFrontmatterPreset = 'yaml' | 'toml';
  interface RemarkFrontmatterMatter {
    type: string;
    marker: string | { open: string; close: string };
    fence?: string | { open: string; close: string };
    anywhere?: boolean;
  }
  type RemarkFrontmatterOptions = Array<
    RemarkFrontmatterPreset | RemarkFrontmatterMatter
  >;

  const remarkFrontmatter: RemarkFrontmatter;
  export = remarkFrontmatter;
}
