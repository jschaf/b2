import { PostAST } from '//post/ast';
import { HastWriter, WriterContext } from '//post/hast/writer';
import * as md from '//post/mdast/nodes';
import * as h from '//post/hast/nodes';
import * as nw from '//post/hast/node_writer';
import * as un from '//unist/nodes';
import { dedent, StringBuilder } from '//strings';
import * as hast from 'hast-format';
import * as unist from 'unist';

const emptyPostAST = PostAST.fromMdast(md.root([]));
const ctx = (ancestors: unist.Node[] = []): WriterContext => {
  const wc = WriterContext.create(emptyPostAST, h.root([]));
  wc.ancestors.push(...ancestors);
  return wc;
};

const indent = (
  literals: TemplateStringsArray,
  ...placeholders: string[]
): string => {
  const d = dedent(literals, ...placeholders);
  return (
    '\n' +
    d
      .split('\n')
      .map(l => '  ' + l)
      .join('\n')
  );
};

describe('CommentWriter', () => {
  it('should write a comment node', () => {
    const sb = StringBuilder.create();
    const w = nw.CommentWriter.create();

    w.writeNode(h.comment('foo'), ctx(), sb);

    expect(sb.toString()).toEqual('<!-- foo -->');
  });
});

describe('DoctypeWriter', () => {
  it('should write a doctype', () => {
    const sb = StringBuilder.create();
    const w = nw.DoctypeWriter.create();

    w.writeNode(h.doctype(), ctx(), sb);

    expect(sb.toString()).toEqual('<!doctype html>\n');
  });
});

describe('ElementWriter', () => {
  const testData: [string, hast.Element, string][] = [
    [
      'div > p',
      h.elem('div', [h.elemText('p', 'foo')]),
      indent`
        <div>
          <p>foo</p>
        </div>
      `,
    ],
    [
      'div[class="a b c" data-foo="qux"}] > p',
      h.elemProps('div', { class: ['a', 'b', 'c'], 'data-foo': 'qux' }, [
        h.elemText('p', 'foo'),
      ]),
      indent`
        <div class="a b c" data-foo="qux">
          <p>foo</p>
        </div>
      `,
    ],
    [
      'unsafe chars',
      h.elemText('code', 'echo 5 < 2 && 3 > 4'),
      `<code>echo 5 &lt; 2 && 3 &gt; 4</code>`,
    ],
  ];
  for (const [name, input, expected] of testData) {
    it(name, () => {
      const sb = StringBuilder.create();
      const c = HastWriter.createDefault();
      const w = nw.ElementWriter.create(c);

      w.writeNode(input, ctx(), sb);

      expect(sb.toString()).toEqual(expected);
    });
  }
  describe('pretty printing', () => {
    const testData: [string, unist.Node, string][] = [
      [
        'body > div > p > text + code',
        h.elem('body', [
          h.elem('div', [
            h.elem('p', [h.text('foo '), h.elemText('code', 'qux')]),
          ]),
        ]),
        dedent`
          <body>
            <div>
              <p>foo <code>qux</code></p>
            </div>
          </body>
        `,
      ],
      [
        'div[class="a b c" data-foo="qux"}] > p',
        h.elem('body', [
          h.elemProps('div', { class: ['a', 'b', 'c'], 'data-foo': 'qux' }, [
            h.elemText('p', 'foo'),
          ]),
        ]),
        dedent`
          <body>
            <div class="a b c" data-foo="qux">
              <p>foo</p>
            </div>
          </body>
        `,
      ],
      [
        'head > meta + link',
        h.elem('head', [
          h.elemProps('meta', { charset: 'foo' }),
          h.elemProps('link', { rel: 'bar' }),
        ]),
        dedent`
          <head>
            <meta charset="foo">
            <link rel="bar">
          </head>
        `,
      ],
    ];
    for (const [name, input, expected] of testData) {
      it(name, () => {
        const sb = StringBuilder.create();
        const c = HastWriter.createDefault();
        const w = nw.ElementWriter.create(c);

        w.writeNode(input, ctx(), sb);
        const actual = sb.toString();

        expect(actual).toEqual('\n' + expected);
      });
    }
  });
});

describe('RawWriter', () => {
  it('should write a raw node', () => {
    const sb = StringBuilder.create();
    const w = nw.RawWriter.create();

    w.writeNode(h.raw('<div>foo</div>'), ctx(), sb);

    expect(sb.toString()).toEqual('<div>foo</div>\n');
  });
});

describe('RootWriter', () => {
  it('should write a root node', () => {
    const sb = StringBuilder.create();
    const c = HastWriter.createDefault();
    const w = nw.RootWriter.create(c);

    w.writeNode(h.root([h.text('foo'), h.raw('<br>')]), ctx(), sb);

    expect(sb.toString()).toEqual('foo<br>\n');
  });
});

describe('TextWriter', () => {
  const testData: [string, unist.Node, WriterContext, string][] = [
    ['foo', un.text('foo'), ctx(), 'foo'],
    [
      'escape < and >',
      un.text('3 < 5 && 7 > 6'),
      ctx(),
      '3 &lt; 5 && 7 &gt; 6',
    ],
    ['escape unsafe ampersands', un.text('3&4;'), ctx(), '3&amp;4;'],
    [
      'not escape script',
      un.text('3<4 && 5<6;'),
      ctx([h.elem('script')]),
      '3<4 && 5<6;',
    ],
    [
      'not escape style',
      un.text('3<4 && 5<6;'),
      ctx([h.elem('style')]),
      '3<4 && 5<6;',
    ],
  ];
  for (const [name, node, ctx, expected] of testData) {
    it(name, () => {
      const sb = StringBuilder.create();
      const w = nw.TextWriter.create();

      w.writeNode(node, ctx, sb);

      expect(sb.toString()).toEqual(expected);
    });
  }
});
