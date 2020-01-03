import * as dates from '//dates';
import { normalizeLabel } from '//post/mdast/nodes';
import * as md from '//post/mdast/nodes';

describe('tomlFrontmatter', () => {
  it('should normalize dates', () => {
    const date = '2019-10-17';
    const ast = md.tomlFrontmatter({ date: dates.fromISO(date) });
    expect(ast).toEqual({ type: 'toml', value: 'date = 2019-10-17' });
  });
});

describe('normalizeLabel', () => {
  const attrs: [string, string, string][] = [
    ['simple', 'a', 'a'],
    ['lowercase', 'AabB', 'aabb'],
    ['leading space', ' \n\t ab', 'ab'],
    ['trailing space', 'ab \t\n', 'ab'],
    ['two spaces', 'a  b', 'a b'],
    ['tab', 'a \t b', 'a b'],
  ];
  for (let [name, input, expected] of attrs) {
    it(`should normalize ${name}, input='${input}'`, () => {
      const s = normalizeLabel(input);
      expect(s).toEqual(expected);
    });
  }
});
