import { PostAST } from '//post/ast';
import * as md from '//post/mdast/nodes';
import * as h from '//post/hast/nodes';
import * as nw from '//post/hast/node_writer';
import * as un from '//unist/nodes';
import { StringBuilder } from '//strings';

const emptyPostAST = PostAST.fromMdast(md.root([]));

describe('DoctypeWriter', () => {
  it('should write a doctype', () => {
    const sb = StringBuilder.create();
    const w = nw.DoctypeWriter.create(sb);

    w.writeNode(h.doctype(), emptyPostAST);

    expect(sb.toString()).toEqual('<!doctype html>\n');
  });
});

describe('RawWriter', () => {
  it('should write a raw node', () => {
    const sb = StringBuilder.create();
    const w = nw.RawWriter.create(sb);

    w.writeNode(h.raw('<div>foo</div>'), emptyPostAST);

    expect(sb.toString()).toEqual('<div>foo</div>\n');
  });
});

describe('TextWriter', () => {
  it('should write a text node', () => {
    const sb = StringBuilder.create();
    const w = nw.TextWriter.create(sb);

    w.writeNode(un.text('foo'), emptyPostAST);

    expect(sb.toString()).toEqual('foo');
  });
});
