import {
  hastElem,
  hastElemText,
  hastElemWithProps, hastRaw,
  hastText,
} from '//post/hast/hast_nodes';
import {MdastCompiler} from '//post/mdast/compiler';
import * as nc from '//post/mdast/node_compiler';
import * as md from '//post/mdast/nodes';
import {PostAST} from '//post/post_ast';
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

    const hast = nc.BlockquoteCompiler.create(c).compileNode(p.mdastNode, p);

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

    const hast = nc.BreakCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual(hastElem('break'));
  });
});

describe('CodeCompiler', () => {
  it('should compile code with a lang', () => {
    let code = 'function foo() {}';
    const p = PostAST.create(md.codeWithLang('javascript', code));

    const hast = nc.CodeCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual(
        hastElem('pre', [
          hastElemWithProps('code', {className: ['lang-javascript']}, [
            hastText(code),
          ]),
        ])
    );
  });

  it('should compile code without a lang', () => {
    let code = 'function foo() {}';
    const post = PostAST.create(md.code(code));

    const hast = nc.CodeCompiler.create().compileNode(post.mdastNode, post);

    expect(hast).toEqual(hastElem('pre', [hastElem('code', [hastText(code)])]));
  });
});

describe('DeleteCompiler', () => {
  it('should compile a delete', () => {
    const p = PostAST.create(
        md.deleted([md.text('first'), md.emphasisText('second')])
    );
    const c = MdastCompiler.createDefault();

    const hast = nc.DeleteCompiler.create(c).compileNode(p.mdastNode, p);

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

    const hast = nc.EmphasisCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual(hastElem('em', [hastText(content)]));
  });
});

describe('FootnoteCompiler', () => {
  it('should compile a footnote', () => {
    const p = PostAST.create(md.footnote([md.text('inline fn')]));
    const c = MdastCompiler.createDefault();

    const hast = nc.FootnoteCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual(
        nc.FootnoteReferenceCompiler.makeHastNode(PostAST.newInlineFootnoteId(1))
    );
  });
});

describe('FootnoteReferenceCompiler', () => {
  it('should compile a footnote reference', () => {
    const id = 'my-fn-ref';
    const p = PostAST.create(md.footnoteRef(id));

    const hast = nc.FootnoteReferenceCompiler.create().compileNode(
        p.mdastNode,
        p
    );

    expect(hast).toEqual(nc.FootnoteReferenceCompiler.makeHastNode(id));
  });
});

describe('HeadingCompiler', () => {
  it('should compile a heading with only text', () => {
    const content = 'foobar';
    const p = PostAST.create(md.heading('h3', [md.text(content)]));
    const c = MdastCompiler.createDefault();

    const hast = nc.HeadingCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual(hastElem('h3', [hastText(content)]));
  });

  it('should compile a heading with other content', () => {
    const p = PostAST.create(
        md.heading('h1', [md.text('start'), md.emphasisText('mid')])
    );
    const c = MdastCompiler.createDefault();

    const hast = nc.HeadingCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual(
        hastElem('h1', [hastText('start'), hastElemText('em', 'mid')])
    );
  });
});

describe('HTMLCompiler', () => {
  it('should compile a html node', () => {
    const a = '<div><alpha></alpha></div>';
    const p = PostAST.create(md.html(a));

    const hast = nc.HTMLCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual(hastRaw(a));
  });
});

describe('InlineCodeCompiler', () => {
  it('should compile inline code', () => {
    const value = 'let a = 2';
    const p = PostAST.create(md.inlineCode(value));

    const hast = nc.InlineCodeCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual(hastElemText('code', value));
  });
});

describe('LinkCompiler', () => {
  it('should compile a link without a title', () => {
    let url = 'www.example.com';
    let value = 'text';
    const p = PostAST.create(md.linkText(url, value));
    const c = MdastCompiler.createDefault();

    const hast = nc.LinkCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual(
        hastElemWithProps('a', {href: url}, [hastText(value)])
    );
  });
});

describe('StrongCompiler', () => {
  it('should compile a strong', () => {
    const a = 'alpha';
    const b = 'bravo';
    const p = PostAST.create(md.strong([md.text(a), md.emphasisText(b)]));
    const c = MdastCompiler.createDefault();

    const hast = nc.StrongCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual(
        hastElem('strong', [hastText(a), hastElemText('em', b)])
    );
  });
});

describe('TomlCompiler', () => {
  it('should ignore toml nodes', () => {
    const p = PostAST.create(md.toml({foo: 'bar'}));

    const hast = nc.TomlCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual(unistNodes.ignored());
  });
});
