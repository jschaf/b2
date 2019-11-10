import {
  checkArg,
  checkDefined,
  checkState,
  PreconditionError,
} from './asserts';

describe('promises', () => {
  describe('checkArg', () => {
    it('should not throw for true values', () => {
      checkArg(true);
      checkArg(2 > 1);
    });

    it('should throw for false expressions', () => {
      expect(() => checkArg(false)).toThrow(PreconditionError);
    });

    it('should throw for false expressions with custom message', () => {
      expect(() => checkArg(false, 'a message')).toThrow(/a message/);
    });
  });

  describe('checkState', () => {
    it('should not throw for true values', () => {
      checkState(true);
      checkState(2 > 1);
    });

    it('should throw for false expressions', () => {
      expect(() => checkState(false)).toThrow(PreconditionError);
    });

    it('should throw for false expressions with custom message', () => {
      expect(() => checkState(false, 'a message')).toThrow(/a message/);
    });
  });

  describe('checkDefinedAndNotNull', () => {
    it('should not throw for a defined value', () => {
      const a = 2;
      checkDefined(a);
    });

    it('should not throw for a defined value that is falsy', () => {
      checkDefined(false);
      checkDefined('');
      checkDefined(0);
    });

    it('should throw for an undefined expressions', () => {
      expect(() => checkDefined(undefined)).toThrow(PreconditionError);
    });

    it('should throw for an undefined expressions with custom message', () => {
      expect(() => checkDefined(undefined, 'custom message')).toThrow(
        /custom message/
      );
    });

    it('should throw for an null expressions', () => {
      expect(() => checkDefined(undefined)).toThrow(PreconditionError);
    });

    it('should throw for an null expressions with custom message', () => {
      expect(() => checkDefined(undefined, 'my message')).toThrow(/my message/);
    });

    it('should narrow a type to non-nullable', () => {
      const maybeTwo = (): number | undefined => 2;
      expect(checkDefined(maybeTwo()) + 1).toBe(3);
    });
  });
});
