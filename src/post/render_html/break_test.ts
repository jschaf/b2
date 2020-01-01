import { BreakRenderer } from '//post/render_html/break';
import { hastElem } from '//post/render_html/hast_nodes';
import { mdBreak } from '//post/testing/markdown_nodes';
import vfile from 'vfile';

describe('BreakRenderer', () => {
  it('should render a break', () => {
    const md = mdBreak();

    const hast = BreakRenderer.create().render(md, vfile());

    expect(hast).toEqual(hastElem('break'));
  });
});
