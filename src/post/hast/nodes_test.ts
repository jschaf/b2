import { danglingImageRef } from '//post/hast/nodes';
import * as md from '//post/mdast/nodes';
import * as h from '//post/hast/nodes';

describe('danglingImageRef', () => {
  const attrs: [md.RefType, string, md.ImageRefProps, string][] = [
    [md.RefType.Full, 'alpha', { alt: 'alt' }, '![alt][alpha]'],
    [md.RefType.Full, 'alpha', { alt: 'alt', label: 'ALPHA' }, '![alt][ALPHA]'],
    [md.RefType.Shortcut, 'alpha', {}, '![alpha]'],
    [md.RefType.Shortcut, 'alpha', { label: 'ALPHA' }, '![ALPHA]'],
    [md.RefType.Collapsed, 'alpha', {}, '![alpha][]'],
    [md.RefType.Collapsed, 'alpha', { label: 'ALPHA' }, '![ALPHA][]'],
  ];
  for (let [ref, id, props, expected] of attrs) {
    it(`should render ref=${ref}, label=${props.label || '<none>'}`, () => {
      const ir = md.imageRefProps(id, ref, props);
      expect(danglingImageRef(ir)).toEqual(h.text(expected));
    });
  }
});
