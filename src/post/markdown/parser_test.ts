import { MarkdownParser } from '//post/markdown/parser';
import * as md from '//post/mdast/nodes';
import * as unistNodes from '//unist/nodes';
import * as unist from 'unist';
import * as frontMatters from '//post/testing/front_matters';

describe('MarkdownParser', () => {
  const yaml = frontMatters.defaultYamlCodeBlock();
  const toml = frontMatters.defaultTomlBlock();
  const mdToml = frontMatters.defaultTomlMdast();
  const h1 = '# hello';
  const mdH1 = md.headingText('h1', 'hello');
  const para = 'Hello world.';
  const mdPara = md.paragraphText(para);

  const testData: [string, string, unist.Node][] = [
    ['empty', '', md.root([])],
    [
      'simple with code frontmatter',
      [h1, yaml, para].join('\n\n'),
      md.root([mdToml, mdH1, mdPara]),
    ],
    [
      'simple with toml frontmatter at front',
      [toml, h1, para].join('\n\n'),
      md.root([mdToml, mdH1, mdPara]),
    ],
    [
      'simple with both yaml and toml frontmatter',
      [toml, h1, yaml, para].join('\n\n'),
      md.root([mdToml, mdH1, mdPara]),
    ],
    [
      'simple with no frontmatter',
      [h1, para].join('\n\n'),
      md.root([mdH1, mdPara]),
    ],
    [
      'paragraph followed by list',
      [para, '1. list item'].join('\n'),
      md.root([mdPara, md.orderedList([md.paragraphText('list item')])]),
    ],
  ];
  for (const [name, input, expected] of testData) {
    it(name, () => {
      const p = MarkdownParser.create();
      const actual = p.parse(input);
      expect(stripPos(actual.mdastNode)).toEqual(expected);
    });
  }
});

const stripPos = (n: unist.Node): unist.Node => {
  unistNodes.removePositionInfo(n);
  return n;
};
