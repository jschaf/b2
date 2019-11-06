import { Mempost } from '//post/mempost';

describe('Mempost', () => {
  it('should equal itself', () => {
    expect(Mempost.ofUtf8Entry('foo', 'bar')).toEqualMempost(
      Mempost.ofUtf8Entry('foo', 'bar')
    );
  });
});
