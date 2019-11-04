import { Mempost } from './mempost';

describe('Mempost', () => {
  it('should equal itself', () => {
    expect(Mempost.ofUtf8Entry('foo', 'bar')).toEqualMempost(
      Mempost.ofUtf8Entry('foo', 'bar')
    );
  });
});
