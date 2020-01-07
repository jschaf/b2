import * as md from '//post/mdast/nodes';
import { PostParser, TEXT_PACK_BUNDLE_PREFIX } from '//post/parser';
import { ZipFileEntry, Zipper } from '//zip_files';
import * as unistNodes from '//unist/nodes';
import * as unist from 'unist';
import * as frontMatters from '//post/testing/front_matters';
import * as mdast from 'mdast';

const stripPos = (n: unist.Node): unist.Node => {
  unistNodes.removePositionInfo(n);
  return n;
};

describe('PostParser', () => {
  const yaml = frontMatters.defaultYamlCodeBlock();
  const toml = frontMatters.defaultTomlBlock();
  const mdToml = frontMatters.defaultTomlMdast();
  const h1 = '# hello';
  const mdH1 = md.headingText('h1', 'hello');
  const para = 'Hello world.';
  const mdPara = md.paragraphText(para);
  const testData: [string, string, mdast.Root][] = [
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
  describe('parseMarkdown', () => {
    for (const [name, markdown, expected] of testData) {
      it(name, () => {
        const p = PostParser.create().parseMarkdown(markdown);
        expect(stripPos(p.mdastNode)).toEqual(expected);
      });
    }
  });

  describe('parseTextPack', () => {
    for (const [name, markdown, expected] of testData) {
      it(name, async () => {
        const buf = await Zipper.zip([
          ZipFileEntry.ofUtf8(TEXT_PACK_BUNDLE_PREFIX + '/text.md', markdown),
        ]);
        const p = await PostParser.create().parseTextPack(buf);
        expect(stripPos(p.mdastNode)).toEqual(expected);
      });
    }
  });
});
