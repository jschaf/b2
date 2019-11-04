import {Mempost} from '../post/mempost';

describe('toEqualMempost', () => {
  it('should equal itself', () => {
    expect(Mempost.ofUtf8Entry('foo', 'bar'))
        .toEqualMempost(Mempost.ofUtf8Entry('foo', 'bar'));
  });

  it('should equal itself with an object', () => {
    expect(Mempost.ofUtf8Entry('foo', 'bar'))
        .toEqualMempost({foo: 'bar'});
  });

  it('should not equal a different Mempost', () => {
    expect(Mempost.ofUtf8Entry('foo', 'bar'))
        .not.toEqualMempost(Mempost.ofUtf8Entry('foo', 'baz'));
  });

  it('should not equal a different object', () => {
    expect(Mempost.ofUtf8Entry('foo', 'bar'))
        .not.toEqualMempost({foo: 'qux'});
  });

  it('should equal differently formatted html', () => {
    expect(Mempost.ofUtf8Entry('foo.html', '<h1>bar</h1>  <p>foo</p>'))
        .toEqualMempost({'foo.html': '<h1>bar</h1><p>foo</p>'});
  });
});
