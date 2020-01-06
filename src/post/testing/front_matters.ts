import * as dates from '//dates';
import { PostMetadata } from '//post/metadata';
import { dedent } from '//strings';
import * as mdast from 'mdast';
import * as md from '//post/mdast/nodes';

export const withDefaultFrontMatter = (text: string): string => {
  const lines = text.split('\n');
  const lineNum = 2;
  lines.splice(lineNum, 0, ...yamlCode.split('\n'));
  return lines.join('\n');
};

const yamlText = dedent`
    # Metadata
    slug: foo_bar
    date: 2019-10-08
`;

const yamlCode = dedent`
    \`\`\`yaml
${yamlText}
    \`\`\`
`;

const tomlText = dedent`
    slug = "foo_bar"
    date = 2019-10-08
`;

export const defaultYamlText = () => yamlText;
export const defaultTomlText = () => tomlText;

export const newCodeMetadata = (value: string): mdast.Code => {
  return md.codeWithLang('yaml', value);
};

export const DEFAULT_FRONTMATTER = PostMetadata.parse({
  slug: 'foo_bar',
  date: dates.fromISO('2019-10-08'),
});
