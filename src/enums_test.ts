import { newTypeGuardCheck } from '//enums';

describe('newTypeGuardCheck', () => {
  it('should work for string enums', () => {
    enum Foo {
      Alpha = 'alpha',
      Bravo = 'bravo',
    }

    const isFoo = newTypeGuardCheck(Foo);

    expect(isFoo('alpha')).toBe(true);
    expect(isFoo(Foo.Alpha)).toBe(true);
    expect(isFoo('bravo')).toBe(true);
    expect(isFoo(Foo.Bravo)).toBe(true);

    expect(isFoo('b')).toBe(false);
    expect(isFoo(1)).toBe(false);
    expect(isFoo(null)).toBe(false);
    expect(isFoo(undefined)).toBe(false);
    expect(isFoo({})).toBe(false);
  });
});
