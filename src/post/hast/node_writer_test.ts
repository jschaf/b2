import { PostAST } from '//post/ast';
import { DoctypeWriter } from '//post/hast/node_writer';
import * as md from '//post/mdast/nodes';
import * as h from '//post/hast/nodes';
import { StringBuilder } from '//strings';

const emptyPostAST = PostAST.fromMdast(md.root([]));

describe('DoctypeWriter', () => {
  it('should write a doctype', () => {
    const sb = StringBuilder.create();
    const w = DoctypeWriter.create(sb);

    w.writeNode(h.doctype(), emptyPostAST);

    expect(sb.toString()).toEqual('<!doctype html>\n');
  });
});
