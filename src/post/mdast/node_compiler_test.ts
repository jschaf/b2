import * as h from '//post/hast/nodes';
import { MdastCompiler } from '//post/mdast/compiler';
import * as nc from '//post/mdast/node_compiler';
import * as md from '//post/mdast/nodes';
import { PostAST } from '//post/post_ast';
import * as mdast from 'mdast';
import * as hast from 'hast-format';
import * as unist from 'unist';

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

    expect(hast).toEqual([
      h.elem('blockquote', [
        h.elemText('p', 'first'),
        h.elem('p', [h.elemText('em', 'second')]),
      ]),
    ]);
  });
});

describe('BreakCompiler', () => {
  it('should compile a break', () => {
    const p = PostAST.create(md.lineBreak());

    const hast = nc.BreakCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual([h.elem('break')]);
  });
});

describe('CodeCompiler', () => {
  it('should compile code with a lang', () => {
    let code = 'function foo() {}';
    const p = PostAST.create(md.codeWithLang('javascript', code));

    const hast = nc.CodeCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual([
      h.elem('pre', [
        h.elemProps('code', { className: ['lang-javascript'] }, [h.text(code)]),
      ]),
    ]);
  });

  it('should compile code without a lang', () => {
    let code = 'function foo() {}';
    const post = PostAST.create(md.code(code));

    const hast = nc.CodeCompiler.create().compileNode(post.mdastNode, post);

    expect(hast).toEqual([h.elem('pre', [h.elem('code', [h.text(code)])])]);
  });
});

describe('DeleteCompiler', () => {
  it('should compile a delete', () => {
    const p = PostAST.create(
      md.deleted([md.text('first'), md.emphasisText('second')])
    );
    const c = MdastCompiler.createDefault();

    const hast = nc.DeleteCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual([
      h.elem('del', [h.text('first'), h.elemText('em', 'second')]),
    ]);
  });
});

describe('EmphasisCompiler', () => {
  it('should compile emphasis with only text', () => {
    const content = 'foobar';
    const p = PostAST.create(md.emphasisText(content));
    const c = MdastCompiler.createDefault();

    const hast = nc.EmphasisCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual([h.elem('em', [h.text(content)])]);
  });
});

describe('FootnoteCompiler', () => {
  it('should compile a footnote', () => {
    const p = PostAST.create(md.footnote([md.text('inline fn')]));
    const c = MdastCompiler.createDefault();

    const hast = nc.FootnoteCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual([
      nc.FootnoteReferenceCompiler.makeHastNode(PostAST.newInlineFootnoteId(1)),
    ]);
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

    expect(hast).toEqual([nc.FootnoteReferenceCompiler.makeHastNode(id)]);
  });
});

describe('HeadingCompiler', () => {
  it('should compile a heading with only text', () => {
    const content = 'foobar';
    const p = PostAST.create(md.heading('h3', [md.text(content)]));
    const c = MdastCompiler.createDefault();

    const hast = nc.HeadingCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual([h.elem('h3', [h.text(content)])]);
  });

  it('should compile a heading with other content', () => {
    const p = PostAST.create(
      md.heading('h1', [md.text('start'), md.emphasisText('mid')])
    );
    const c = MdastCompiler.createDefault();

    const hast = nc.HeadingCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual([
      h.elem('h1', [h.text('start'), h.elemText('em', 'mid')]),
    ]);
  });
});

describe('HTMLCompiler', () => {
  it('should compile a html node', () => {
    const a = '<div><alpha></alpha></div>';
    const p = PostAST.create(md.html(a));

    const hast = nc.HTMLCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual([h.raw(a)]);
  });
});

describe('ImageCompiler', () => {
  it('should compile an image without any props', () => {
    const src = 'http://example.com';
    const p = PostAST.create(md.image(src));

    const hast = nc.ImageCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual([h.elemProps('img', { src })]);
  });

  it('should compile an image with a title and alt attr', () => {
    const src = 'http://example.com';
    const title = 'my title';
    const alt = 'alt text';
    const p = PostAST.create(md.imageProps(src, { title, alt }));

    const hast = nc.ImageCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual([h.elemProps('img', { src, title, alt })]);
  });
});

describe('ImageReferenceCompiler', () => {
  it('should compile a dangling image reference node (full)', () => {
    let imgRef = md.imageRefProps('alpha', md.RefType.Full, {
      alt: 'alt',
    });
    const p = PostAST.create(imgRef);
    const c = MdastCompiler.createDefault();

    const hast = nc.ImageReferenceCompiler.create(c).compileNode(
      p.mdastNode,
      p
    );

    expect(hast).toEqual([h.danglingImageRef(imgRef)]);
  });

  it('should compile an image reference node', () => {
    const id = 'alpha';
    const alt = 'alt';
    const title = 'title';
    const src = 'http://bravo.com';
    let imgRef = md.imageRefProps(id, md.RefType.Full, { alt });
    const p = PostAST.create(imgRef);
    p.addDefinition(md.definitionProps(id, src, { title }));
    const c = MdastCompiler.createDefault();

    const hast = nc.ImageReferenceCompiler.create(c).compileNode(
      p.mdastNode,
      p
    );

    expect(hast).toEqual([h.elemProps('img', { src, title, alt })]);
  });
});

describe('InlineCodeCompiler', () => {
  it('should compile inline code', () => {
    const value = 'let a = 2';
    const p = PostAST.create(md.inlineCode(value));

    const hast = nc.InlineCodeCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual([h.elemText('code', value)]);
  });
});

describe('LinkCompiler', () => {
  it('should compile a link without a title', () => {
    let url = 'www.example.com';
    let value = 'text';
    const p = PostAST.create(md.linkText(url, value));
    const c = MdastCompiler.createDefault();

    const hast = nc.LinkCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual([h.elemProps('a', { href: url }, [h.text(value)])]);
  });
});

describe('LinkReferenceCompiler', () => {
  it('should compile a dangling link reference node', () => {
    const id = 'alpha';
    const text = 'bravo';
    let lr = md.linkRefText(id, md.RefType.Full, text);
    const p = PostAST.create(lr);
    const c = MdastCompiler.createDefault();
    let childrenCompiler = (n: mdast.LinkReference) => c.compileChildren(n, p);

    const hast = nc.LinkReferenceCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual(h.danglingLinkRef(lr, childrenCompiler));
  });

  it('should compile a link reference', () => {
    const id = 'alpha';
    const url = 'http://example';
    let lr = md.linkRef(id, md.RefType.Full, [
      md.emphasisText('foo'),
      md.text('bar'),
    ]);
    const p = PostAST.create(lr);
    p.addDefinition(md.definitionProps(id, url, { title: 'title' }));
    const c = MdastCompiler.createDefault();

    const hast = nc.LinkReferenceCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual([
      h.elemProps('a', { href: url, title: 'title' }, [
        h.elemText('em', 'foo'),
        h.text('bar'),
      ]),
    ]);
  });
});

describe('ListCompiler', () => {
  const a = 'alpha';
  const b = 'bravo';
  const spreadListItem = (cs: mdast.BlockContent[]) =>
    md.listItemProps({ spread: true }, cs);
  const para = md.paragraphText;
  const pTag = (value: string) => h.elemText('p', value);
  const checkbox = nc.ListItemCompiler.checkbox;

  const testData: [string, mdast.List, hast.Element][] = [
    [
      'unordered, tight, 1 item',
      md.list([md.listItemText(b)]),
      h.elem('ul', [h.elemText('li', b)]),
    ],
    [
      'ordered, loose, 2 items',
      md.listProps({ spread: true, ordered: true }, [
        spreadListItem([para(a)]),
        spreadListItem([para(b)]),
      ]),
      h.elem('ol', [h.elem('li', [pTag(a)]), h.elem('li', [pTag(b)])]),
    ],
    [
      'ordered, tight, some checkboxes',
      md.listProps({ ordered: true }, [
        md.listItemProps({ checked: true }, [para(a)]),
        md.listItemProps({ checked: false }, [para(b)]),
      ]),
      h.elemProps('ol', { className: [nc.ListCompiler.CHECKBOX_CLASS_NAME] }, [
        h.elem('li', [checkbox(true), h.text(a)]),
        h.elem('li', [checkbox(false), h.text(b)]),
      ]),
    ],
  ];

  for (const [name, input, expected] of testData) {
    it(`should compile ${name}`, () => {
      const p = PostAST.create(input);
      const c = MdastCompiler.createDefault();

      const hast = nc.ListCompiler.create(c).compileNode(p.mdastNode, p);

      expect(hast).toEqual([expected]);
    });
  }
});

describe('ListItemCompiler', () => {
  const a = 'alpha';
  const b = 'bravo';
  const para = md.paragraphText;
  const pTag = (value: string) => h.elemText('p', value);
  const t = h.text;
  const checkbox = nc.ListItemCompiler.checkbox;
  type Data = [string, md.ListItemProps, mdast.BlockContent[], unist.Node[]];
  const testData: Data[] = [
    ['tight, not checkbox', {}, [para(b)], [t(b)]],
    ['tight, checked', { checked: true }, [para(a)], [checkbox(true), t(a)]],
    [
      'tight, unchecked',
      { checked: false },
      [para(a)],
      [checkbox(false), t(a)],
    ],
    ['loose, not checkbox', { spread: true }, [para(b)], [pTag(b)]],
    [
      'loose, checked',
      { spread: true, checked: true },
      [para(a)],
      [checkbox(true), pTag(a)],
    ],
    [
      'loose, unchecked',
      { spread: true, checked: false },
      [para(a)],
      [checkbox(false), pTag(a)],
    ],
    ['tight, multiple para', {}, [para(a), para(b)], [t(a), t(b)]],
  ];

  for (const [name, props, input, expected] of testData) {
    it(`should compile ${name} list`, () => {
      const p = PostAST.create(md.listItemProps(props, input));
      const c = MdastCompiler.createDefault();

      const hast = nc.ListItemCompiler.create(c).compileNode(p.mdastNode, p);

      expect(hast).toEqual([h.elem('li', expected)]);
    });
  }
});

describe('StrongCompiler', () => {
  it('should compile a strong', () => {
    const a = 'alpha';
    const b = 'bravo';
    const p = PostAST.create(md.strong([md.text(a), md.emphasisText(b)]));
    const c = MdastCompiler.createDefault();

    const hast = nc.StrongCompiler.create(c).compileNode(p.mdastNode, p);

    expect(hast).toEqual([h.elem('strong', [h.text(a), h.elemText('em', b)])]);
  });
});

describe('TomlCompiler', () => {
  it('should ignore toml nodes', () => {
    const p = PostAST.create(md.toml({ foo: 'bar' }));

    const hast = nc.TomlCompiler.create().compileNode(p.mdastNode, p);

    expect(hast).toEqual([]);
  });
});
