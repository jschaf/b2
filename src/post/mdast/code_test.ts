import { CodeCompiler } from '//post/mdast/code';
import { hastElem, hastElemWithProps, hastText } from '//post/mdast/hast_nodes';
import { PostAST } from '//post/post_ast';
import { mdCode, mdCodeWithLang } from '//post/testing/markdown_nodes';

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
