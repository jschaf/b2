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

const yamlCode = ['```yaml', yamlText, '```'].join('\n');

const tomlText = dedent`
    slug = "foo_bar"
    date = 2019-10-08
`;
const tomlBlock = ['+++', tomlText, '+++'].join('\n');

export const defaultYamlText = () => yamlText;
export const defaultYamlCodeBlock = () => yamlCode;

export const defaultTomlText = () => tomlText;
export const defaultTomlBlock = () => tomlBlock;
export const defaultTomlMdast = () => md.tomlText(tomlText);

export const newCodeMetadata = (value: string): mdast.Code => {
  return md.codeWithLang('yaml', value);
};

export const DEFAULT_FRONTMATTER = PostMetadata.parse({
  slug: 'foo_bar',
  date: dates.fromISO('2019-10-08'),
});
