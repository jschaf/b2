import {
  hastElem,
  hastElemText,
  hastElemWithProps,
  hastText,
} from '//post/hast/hast_nodes';
import { MdastCompiler } from '//post/mdast/compiler';
import {
  BlockquoteCompiler,
  BreakCompiler,
  CodeCompiler,
  DeleteCompiler,
  EmphasisCompiler,
  FootnoteCompiler,
  FootnoteReferenceCompiler,
  HeadingCompiler,
  InlineCodeCompiler,
  LinkCompiler,
  TomlCompiler,
} from '//post/mdast/node_compiler';
import * as md from '//post/mdast/nodes';
import { PostAST } from '//post/post_ast';
import * as unistNodes from '//unist/nodes';

describe('BlockquoteCompiler', () => {
  it('should compile a blockquote', () => {
    const p = PostAST.create(
      md.blockquote([
        md.paragraphText('first'),
        md.paragraph([md.emphasisText('second')]),
      ])
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
    const p = PostAST.create(md.lineBreak());

    const hast = BreakCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual(hastElem('break'));
  });
});

describe('CodeCompiler', () => {
  it('should compile code with a lang', () => {
    let code = 'function foo() {}';
    const p = PostAST.create(md.codeWithLang('javascript', code));

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
    const post = PostAST.create(md.code(code));

    const hast = CodeCompiler.create().compileNode(post.mdastNode, post);

    expect(hast).toEqual(hastElem('pre', [hastElem('code', [hastText(code)])]));
  });
});

describe('DeleteCompiler', () => {
  it('should compile a delete', () => {
    const p = PostAST.create(
      md.deleted([md.text('first'), md.emphasisText('second')])
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
    const p = PostAST.create(md.emphasisText(content));
    const c = MdastCompiler.createDefault();

    const hast = EmphasisCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual(hastElem('em', [hastText(content)]));
  });
});

describe('FootnoteCompiler', () => {
  it('should compile a footnote', () => {
    const p = PostAST.create(md.footnote([md.text('inline fn')]));
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
    const p = PostAST.create(md.footnoteRef(id));

    const hast = FootnoteReferenceCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual(FootnoteReferenceCompiler.makeHastNode(id));
  });
});

describe('HeadingCompiler', () => {
  it('should compile a heading with only text', () => {
    const content = 'foobar';
    const p = PostAST.create(md.heading('h3', [md.text(content)]));
    const c = MdastCompiler.createDefault();

    const hast = HeadingCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual(hastElem('h3', [hastText(content)]));
  });

  it('should compile a heading with other content', () => {
    const p = PostAST.create(
      md.heading('h1', [md.text('start'), md.emphasisText('mid')])
    );
    const c = MdastCompiler.createDefault();

    const hast = HeadingCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual(
      hastElem('h1', [hastText('start'), hastElemText('em', 'mid')])
    );
  });
});

describe('InlineCodeCompiler', () => {
  it('should compile inline code', () => {
    const value = 'let a = 2';
    const p = PostAST.create(md.inlineCode(value));

    const hast = InlineCodeCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual(hastElemText('code', value));
  });
});

describe('LinkCompiler', () => {
  it('should compile a link without a title', () => {
    let url = 'www.example.com';
    let value = 'text';
    const p = PostAST.create(md.linkText(url, value));
    const c = MdastCompiler.createDefault();

    const hast = LinkCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual(
      hastElemWithProps('a', { href: url }, [hastText(value)])
    );
  });
});

describe('TomlCompiler', () => {
  it('should ignore toml nodes', () => {
    const p = PostAST.create(md.toml({ foo: 'bar' }));

    const hast = TomlCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual(unistNodes.ignored());
  });
});
