import { MdastCompiler } from '//post/mdast/compiler';
import {hastElem, hastElemText, hastElemWithProps, hastText} from '//post/hast/hast_nodes';
import {BlockquoteCompiler, BreakCompiler, CodeCompiler, DeleteCompiler, EmphasisCompiler, FootnoteCompiler, FootnoteReferenceCompiler, HeadingCompiler} from '//post/mdast/node_compiler';
import { PostAST } from '//post/post_ast';
import {
  mdBlockquote, mdBreak, mdCode, mdCodeWithLang, mdDelete,
  mdEmphasisText, mdFootnoteRef, mdHeading, mdInlineFootnote,
  mdPara,
  mdParaText, mdText,
} from '//post/testing/markdown_nodes';

describe('BlockquoteCompiler', () => {
  it('should compile a blockquote', () => {
    const p = PostAST.create(
        mdBlockquote([mdParaText('first'), mdPara([mdEmphasisText('second')])])
    );
    const c = MdastCompiler.createDefault();

    const hast = BlockquoteCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual(
        hastElem('blockquote', [
          hastElemText('p', 'first'),
          hastElem('p', [hastElemText('em', 'second')]),
        ])
    );
  });
});

describe('BreakCompiler', () => {
  it('should compile a break', () => {
    const p = PostAST.create(mdBreak());

    const hast = BreakCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual(hastElem('break'));
  });
});

describe('CodeCompiler', () => {
  it('should compile code with a lang', () => {
    let code = 'function foo() {}';
    const p = PostAST.create(mdCodeWithLang('javascript', code));

    const hast = CodeCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual(
        hastElem('pre', [
          hastElemWithProps('code', { className: ['lang-javascript'] }, [
            hastText(code),
          ]),
        ])
    );
  });

  it('should compile code without a lang', () => {
    let code = 'function foo() {}';
    const post = PostAST.create(mdCode(code));

    const hast = CodeCompiler.create().compileNode(post.mdastNode, post);

    expect(hast).toEqual(hastElem('pre', [hastElem('code', [hastText(code)])]));
  });
});

describe('DeleteCompiler', () => {
  it('should compile a delete', () => {
    const p = PostAST.create(
        mdDelete([mdText('first'), mdEmphasisText('second')])
    );
    const c = MdastCompiler.createDefault();

    const hast = DeleteCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual(
        hastElem('del', [hastText('first'), hastElemText('em', 'second')])
    );
  });
});

describe('EmphasisCompiler', () => {
  it('should compile emphasis with only text', () => {
    const content = 'foobar';
    const p = PostAST.create(mdEmphasisText(content));
    const c = MdastCompiler.createDefault();

    const hast = EmphasisCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual(hastElem('em', [hastText(content)]));
  });
});

describe('FootnoteCompiler', () => {
  it('should compile a footnote', () => {
    const p = PostAST.create(mdInlineFootnote([mdText('inline fn')]));
    const c = MdastCompiler.createDefault();

    const hast = FootnoteCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual(
        FootnoteReferenceCompiler.makeHastNode(PostAST.newInlineFootnoteId(1))
    );
  });
});

describe('FootnoteReferenceCompiler', () => {
  it('should compile a footnote reference', () => {
    const id = 'my-fn-ref';
    const p = PostAST.create(mdFootnoteRef(id));

    const hast = FootnoteReferenceCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual(FootnoteReferenceCompiler.makeHastNode(id));
  });
});

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

