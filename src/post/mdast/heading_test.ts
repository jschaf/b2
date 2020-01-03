import { MdastCompiler } from '//post/mdast/compiler';
import { hastElem, hastElemText, hastText } from '//post/mdast/hast_nodes';
import { HeadingCompiler } from '//post/mdast/heading';
import { PostAST } from '//post/post_ast';
import {
  mdEmphasisText,
  mdHeading,
  mdText,
} from '//post/testing/markdown_nodes';

describe('HeadingCompiler', () => {
  it('should compile a heading with only text', () => {
    const content = 'foobar';
    const p = PostAST.create(mdHeading('h3', [mdText(content)]));
    const c = MdastCompiler.createDefault();

    const hast = HeadingCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual(hastElem('h3', [hastText(content)]));
  });

  it('should compile a heading with other content', () => {
    const p = PostAST.create(
      mdHeading('h1', [mdText('start'), mdEmphasisText('mid')])
    );
    const c = MdastCompiler.createDefault();

    const hast = HeadingCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual(
      hastElem('h1', [hastText('start'), hastElemText('em', 'mid')])
    );
  });
});
