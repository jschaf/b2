import { checkDefined } from '//asserts';
import * as md from '//post/mdast/nodes';
import { PostMetadata } from '//post/metadata';
import * as frontMatters from '//post/testing/front_matters';
import * as dates from '//dates';
import { dedent } from '//strings';
import * as mdast from 'mdast';

test('parses valid tokens', () => {
  const slug = 'qux_bar';
  let date = '2017-06-18';
  const tree = md.root([
    md.headingText('h1', 'hello'),
    md.code(dedent`
        # Metadata
        slug: ${slug}
        date: ${date}
      `),
  ]);

  const metadata = checkDefined(PostMetadata.parseFromMdast(tree));

  expect(metadata.schema).toEqual({ slug, date: dates.fromISO(date) });
});

test('parses valid tokens from toml frontmatter', () => {
  const slug = 'qux_bar';
  let date = dates.fromISO('2019-05-18');
  const tree = md.root([md.tomlFrontmatter({ slug, date })]);

  const metadata = checkDefined(PostMetadata.parseFromMdast(tree));

  expect(metadata.schema).toEqual({ slug, date });
});

describe('normalizeMdast', () => {
  const h1 = md.headingText('h1', 'alpha');
  const toml = md.tomlText(frontMatters.defaultTomlText());
  const yaml = frontMatters.newCodeMetadata(frontMatters.defaultYamlText());
  const input = md.root([h1, toml]);
  const expected = md.root([toml, h1]);

  const testData: [string, mdast.Content[], mdast.Content[]][] = [
    ['h1, yaml, toml', [h1, yaml, toml], [toml, h1]],
    ['h1, toml', [h1, toml], [toml, h1]],
    ['h1, yaml', [h1, yaml], [toml, h1]],
    ['h1', [h1], [h1]],
  ];
  for (const [name, input, expected] of testData) {
    it(name, () => {
      const p = PostMetadata.normalizeMdast(md.root(input));
      expect(p).toEqual(md.root(expected));
    });
  }

  it('should normalize', () => {
    const actual = PostMetadata.normalizeMdast(input);

    expect(actual).toEqual(expected);
  });
});
