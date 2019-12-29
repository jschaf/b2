import {isObject, lossyClone} from '//objects';

describe('isObject', () => {
  it('should work', () => {
    expect(isObject(null)).toBe(false);
    expect(isObject(1)).toBe(false);
    expect(isObject('')).toBe(false);
    expect(isObject('aa')).toBe(false);
    expect(isObject(undefined)).toBe(false);

    expect(isObject({})).toBe(true);
    expect(isObject({a: 1})).toBe(true);
  });
});

describe('lossyClone', () => {
  it('should work', () => {
    expect(lossyClone({})).toEqual({});
    expect(lossyClone({a: 1, b: 2})).toEqual({a: 1, b: 2});
    expect(lossyClone({b: "foo"})).toEqual({b: "foo"});
    expect(lossyClone({a: 1.2})).toEqual({a: 1.2});
    expect(lossyClone({a: {b: 'c'}})).toEqual({a: {b: 'c'}});
  });
});

