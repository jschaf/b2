import { CodeRenderer } from '//post/render_html/code';
import {
  hastElem,
  hastElemWithProps,
  hastText,
} from '//post/render_html/hast_nodes';
import { mdCode, mdCodeWithLang } from '//post/testing/markdown_nodes';
import vfile from 'vfile';

describe('CodeRenderer', () => {
  it('should render code with a lang', () => {
    let code = 'function foo() {}';
    const md = mdCodeWithLang('javascript', code);

    const hast = CodeRenderer.create().render(md, vfile());

    expect(hast).toEqual(
      hastElem('pre', [
        hastElemWithProps('code', { className: ['lang-javascript'] }, [
          hastText(code),
        ]),
      ])
    );
  });

  it('should render code without a lang', () => {
    let code = 'function foo() {}';
    const md = mdCode(code);

    const hast = CodeRenderer.create().render(md, vfile());

    expect(hast).toEqual(hastElem('pre', [hastElem('code', [hastText(code)])]));
  });
});
