import { HastCompiler } from '//post/hast/compiler';
import * as h from '//post/hast/nodes';

describe('HastCompiler', () => {
  it('should compile body > p', () => {
    const a = h.elem('body', [h.elemText('p', 'foo bar')]);

    const html = HastCompiler.create().compile(a);

    expect(html).toEqual('<body><p>foo bar</p></body>');
  });
});
