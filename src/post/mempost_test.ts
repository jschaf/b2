import {Mempost} from '//post/mempost';

describe('Mempost', () => {
  const alpha = 'alpha';
  const bravo = 'bravo';
  const charlie = Buffer.from('charlie');
  const delta = Buffer.from('delta');

  it('should equal itself', () => {
    expect(Mempost.ofUtf8Entry('foo', 'bar')).toEqualMempost(
      Mempost.ofUtf8Entry('foo', 'bar')
    );
  });

  it('should get entries', () => {
    const mp = Mempost.create();
    mp.addUtf8Entry('a', alpha);
    mp.addEntry('c', charlie);

    expect(mp.getUtf8Entry('a')).toEqual(alpha);
    expect(mp.getEntry('c')).toEqual(charlie);
  });

  it('should list all entries', () => {
    const mp = Mempost.create();
    mp.addUtf8Entry('a', alpha);
    mp.addUtf8Entry('b', bravo);
    mp.addEntry('c', charlie);
    mp.addEntry('d', delta);
    const results = new Map<string, string | Buffer>();

    for (const [path, content] of mp.entries()) {
      results.set(path, content);
    }

    const expected: [string, string | Buffer][] = [
      ['a', alpha],
      ['b', bravo],
      ['c', charlie],
      ['d', delta],
    ];
    expect(results).toEqual(new Map(expected))
  });
});
