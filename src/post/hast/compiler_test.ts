import { HastCompiler } from '//post/hast/compiler';
import { hastElem, hastElemText } from '//post/hast/hast_nodes';

describe('HastCompiler', () => {
  it('should compile body > p', () => {
    const h = hastElem('body', [hastElemText('p', 'foo bar')]);

    const html = HastCompiler.create().compile(h);

    expect(html).toEqual('<body><p>foo bar</p></body>');
  });
});
