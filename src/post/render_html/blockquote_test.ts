import { BlockquoteRenderer } from '//post/render_html/blockquote';
import { RenderDispatcher } from '//post/render_html/dispatch';
import { hastElem, hastElemText } from '//post/render_html/hast_nodes';
import {
  mdBlockquote,
  mdEmphasisText,
  mdPara,
  mdParaText,
} from '//post/testing/markdown_nodes';
import vfile from 'vfile';

describe('BlockquoteRenderer', () => {
  it('should render a blockquote', () => {
    const md = mdBlockquote([
      mdParaText('first'),
      mdPara([mdEmphasisText('second')]),
    ]);
    const rd = RenderDispatcher.defaultInstance();

    const hast = BlockquoteRenderer.create(rd).render(md, vfile());

    expect(hast).toEqual(
      hastElem('blockquote', [
        hastElemText('p', 'first'),
        hastElem('p', [hastElemText('em', 'second')]),
      ])
    );
  });
});
