import { PostAST } from '//post/ast';
import { DocTemplate } from '//post/hast/doc_template';
import { HastWriter, WriterContext } from '//post/hast/writer';
import * as h from '//post/hast/nodes';
import * as md from '//post/mdast/nodes';
import { dedent, StringBuilder } from '//strings';
import * as unist from 'unist';

describe('HastWriter', () => {
  it('should compile body > p', () => {
    const ast = PostAST.fromMdast(md.root([]));
    const a = h.elem('body', [h.elemText('p', 'foo bar')]);

    const html = HastWriter.createDefault().write(a, ast);

    expect(html).toMatchSnapshot();
  });

  describe('formatting', () => {
    const testData: [string, unist.Node, string][] = [
      [
        'html > meta + meta + link + script',
        DocTemplate.create()
          .addToHead(
            h.elemProps('meta', { charset: 'foo' }),
            h.elemProps('meta', { charset: 'bar' }),
            h.elemProps('link', { rel: 'icon', href: '/favicon.ico' }),
            h.elemProps('script', {
              defer: true,
              src: '/baz.js',
              type: 'module',
            })
          )
          .render([h.elem('body', [h.elemText('p', 'foo')])]),
        dedent`
          <!doctype html>
          <html lang="en">
          <head>
            <meta charset="foo">
            <meta charset="bar">
            <link rel="icon" href="/favicon.ico">
            <script defer src="/baz.js" type="module"></script>
          </head>
          <body>
            <p>foo</p>
          </body>
          </html>
        `,
      ],
    ];

    const emptyPostAST = () => PostAST.fromMdast(md.root([]));
    for (const [name, input, expected] of testData) {
      it(name, () => {
        const sb = StringBuilder.create();
        const c = HastWriter.createDefault();
        const ctx = WriterContext.create(emptyPostAST());

        c.writeNode(input, ctx, sb);
        const actual = sb.toString();

        expect(actual).toEqual(expected);
      });
    }
  });
});
