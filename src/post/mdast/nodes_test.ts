import * as dates from '//dates';
import * as md from '//post/mdast/nodes';

describe('tomlFrontmatter', () => {
  it('should normalize dates', () => {
    const date = '2019-10-17';
    const ast = md.tomlFrontmatter({ date: dates.fromISO(date) });
    expect(ast).toEqual({ type: 'toml', value: 'date = 2019-10-17' });
  });
});
