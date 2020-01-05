import * as dates from '//dates';
import { PostMetadata } from '//post/metadata';
import { dedent } from '//strings';

export const withDefaultFrontMatter = (text: string): string => {
  const lines = text.split('\n');
  const lineNum = 2;
  lines.splice(lineNum, 0, ...DEFAULT_FRONTMATTER_TEXT.split('\n'));
  return lines.join('\n');
};

export const DEFAULT_FRONTMATTER_TEXT = dedent`
    \`\`\`yaml
    # Metadata
    slug: foo_bar
    date: 2019-10-08
    \`\`\`
`;

export const DEFAULT_FRONTMATTER = PostMetadata.parse({
  slug: 'foo_bar',
  date: dates.fromISO('2019-10-08'),
});
