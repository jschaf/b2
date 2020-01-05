import * as dates from '//dates';
import { normalizeLabel } from '//post/mdast/nodes';
import * as md from '//post/mdast/nodes';
import * as mdast from 'mdast';

describe('table', () => {
  const a = 'alpha';
  const b = 'bravo';
  const tbl = (cs: mdast.TableContent[]) => md.tableProps({}, cs);
  const row = md.tableRow;
  const cell = md.tableCell;
  const t = md.text;
  type Data = [string, md.TableShortcutRow[], mdast.Table];

  const testData: Data[] = [
    [
      'shortcut rows and 1-elem cells',
      [[t(a)], [t(b)]],
      tbl([row([cell([t(a)])]), row([cell([t(b)])])]),
    ],
    ['2-elem cells', [[[t(a), t(a)]]], tbl([row([cell([t(a), t(a)])])])],
    [
      'mix of shortcuts and full with 1-elem cells',
      [[md.tableCellText(a)], row([cell([t(b)])])],
      tbl([row([cell([t(a)])]), row([cell([t(b)])])]),
    ],
  ];
  for (const [name, input, expected] of testData) {
    it(`${name}`, () => {
      const t = md.table(input);
      expect(t).toEqual(expected);
    });
  }

  it('should throw for an uneven table', () => {
    expect(() => md.table([[], [t(b)]])).toThrow(/Uneven/);
    expect(() => md.table([[t(a), t(a)], [t(b)]])).toThrow(/Uneven/);
  });
});

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
