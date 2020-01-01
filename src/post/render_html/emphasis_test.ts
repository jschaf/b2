import { RenderDispatcher } from '//post/render_html/dispatch';
import { EmphasisRenderer } from '//post/render_html/emphasis';
import { hastElem, hastText } from '//post/render_html/hast_nodes';
import { mdEmphasisText } from '//post/testing/markdown_nodes';
import vfile from 'vfile';

describe('EmphasisRenderer', () => {
  it('should render emphasis with only text', () => {
    const content = 'foobar';
    const md = mdEmphasisText(content);
    const rd = RenderDispatcher.defaultInstance();

    const hast = EmphasisRenderer.create(rd).render(md, vfile());

    expect(hast).toEqual(hastElem('em', [hastText(content)]));
  });
});
