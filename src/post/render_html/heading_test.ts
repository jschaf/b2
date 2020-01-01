import { RenderDispatcher } from '//post/render_html/dispatch';
import { hastElem, hastText } from '//post/render_html/hast_nodes';
import { HeadingRenderer } from '//post/render_html/heading';
import {
  mdEmphasisText,
  mdHeading,
  mdText,
} from '//post/testing/markdown_nodes';
import vfile from 'vfile';

describe('HeadingRenderer', () => {
  it('should render a heading with only text', () => {
    const content = 'foobar';
    const md = mdHeading('h3', [mdText(content)]);
    const rd = RenderDispatcher.defaultInstance();

    const hast = HeadingRenderer.create(rd).render(md, vfile());

    expect(hast).toEqual(hastElem('h3', [hastText(content)]));
  });

  it('should render a heading with other content', () => {
    const md = mdHeading('h1', [mdText('start'), mdEmphasisText('mid')]);
    const rd = RenderDispatcher.defaultInstance();

    const hast = HeadingRenderer.create(rd).render(md, vfile());

    expect(hast).toEqual(
      hastElem('h1', [hastText('start'), hastElem('em', [hastText('mid')])])
    );
  });
});
