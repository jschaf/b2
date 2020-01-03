import * as h from '//post/hast/nodes';
import { MdastCompiler } from '//post/mdast/compiler';
import * as nc from '//post/mdast/node_compiler';
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

    const hast = nc.BlockquoteCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual(
      h.elem('blockquote', [
        h.elemText('p', 'first'),
        h.elem('p', [h.elemText('em', 'second')]),
      ])
    );
  });
});

describe('BreakCompiler', () => {
  it('should compile a break', () => {
    const p = PostAST.create(md.lineBreak());

    const hast = nc.BreakCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual(h.elem('break'));
  });
});

describe('CodeCompiler', () => {
  it('should compile code with a lang', () => {
    let code = 'function foo() {}';
    const p = PostAST.create(md.codeWithLang('javascript', code));

    const hast = nc.CodeCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual(
      h.elem('pre', [
        h.elemProps('code', { className: ['lang-javascript'] }, [h.text(code)]),
      ])
    );
  });

  it('should compile code without a lang', () => {
    let code = 'function foo() {}';
    const post = PostAST.create(md.code(code));

    const hast = nc.CodeCompiler.create().compileNode(post.mdastNode, post);

    expect(hast).toEqual(h.elem('pre', [h.elem('code', [h.text(code)])]));
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
      h.elem('del', [h.text('first'), h.elemText('em', 'second')])
    );
  });
});

describe('EmphasisCompiler', () => {
  it('should compile emphasis with only text', () => {
    const content = 'foobar';
    const p = PostAST.create(md.emphasisText(content));
    const c = MdastCompiler.createDefault();

    const hast = nc.EmphasisCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual(h.elem('em', [h.text(content)]));
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

    expect(hast).toEqual(h.elem('h3', [h.text(content)]));
  });

  it('should compile a heading with other content', () => {
    const p = PostAST.create(
      md.heading('h1', [md.text('start'), md.emphasisText('mid')])
    );
    const c = MdastCompiler.createDefault();

    const hast = nc.HeadingCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual(
      h.elem('h1', [h.text('start'), h.elemText('em', 'mid')])
    );
  });
});

describe('HTMLCompiler', () => {
  it('should compile a html node', () => {
    const a = '<div><alpha></alpha></div>';
    const p = PostAST.create(md.html(a));

    const hast = nc.HTMLCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual(h.raw(a));
  });
});

describe('ImageCompiler', () => {
  it('should compile an image without any props', () => {
    const src = 'http://example.com';
    const p = PostAST.create(md.image(src));

    const hast = nc.ImageCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual(h.elemProps('img', { src }));
  });

  it('should compile an image with a title and alt attr', () => {
    const src = 'http://example.com';
    const title = 'my title';
    const alt = 'alt text';
    const p = PostAST.create(md.imageProps(src, { title, alt }));

    const hast = nc.ImageCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual(h.elemProps('img', { src, title, alt }));
  });
});

describe('ImageReferenceCompiler', () => {
  it('should compile a dangling image reference node', () => {
    let imgRef = md.imageRefProps('alpha', md.ReferenceType.Full, {
      alt: 'alt',
    });
    const p = PostAST.create(imgRef);
    const c = MdastCompiler.createDefault();

    const hast = nc.ImageReferenceCompiler.create(c).compileNode(
      p.mdastNode,
      p
    );

    expect(hast).toEqual(h.danglingImageRef(imgRef));
  });

  it('should compile an image reference node', () => {
    const id = 'alpha';
    const alt = 'alt';
    const title = 'title';
    const src = 'http://bravo.com';
    let imgRef = md.imageRefProps(id, md.ReferenceType.Full, { alt });
    const p = PostAST.create(imgRef);
    p.defsById.set(id, md.definitionProps(id, src, { title }));
    const c = MdastCompiler.createDefault();

    const hast = nc.ImageReferenceCompiler.create(c).compileNode(
      p.mdastNode,
      p
    );

    expect(hast).toEqual(h.elemProps('img', { src, title, alt }));
  });
});

describe('InlineCodeCompiler', () => {
  it('should compile inline code', () => {
    const value = 'let a = 2';
    const p = PostAST.create(md.inlineCode(value));

    const hast = nc.InlineCodeCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual(h.elemText('code', value));
  });
});

describe('LinkCompiler', () => {
  it('should compile a link without a title', () => {
    let url = 'www.example.com';
    let value = 'text';
    const p = PostAST.create(md.linkText(url, value));
    const c = MdastCompiler.createDefault();

    const hast = nc.LinkCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual(h.elemProps('a', { href: url }, [h.text(value)]));
  });
});

describe('StrongCompiler', () => {
  it('should compile a strong', () => {
    const a = 'alpha';
    const b = 'bravo';
    const p = PostAST.create(md.strong([md.text(a), md.emphasisText(b)]));
    const c = MdastCompiler.createDefault();

    const hast = nc.StrongCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual(h.elem('strong', [h.text(a), h.elemText('em', b)]));
  });
});

describe('TomlCompiler', () => {
  it('should ignore toml nodes', () => {
    const p = PostAST.create(md.toml({ foo: 'bar' }));

    const hast = nc.TomlCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual(unistNodes.ignored());
  });
});
