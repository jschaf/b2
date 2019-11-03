import {checkArg, checkDefinedAndNotNull, checkState, PreconditionError} from "./asserts";

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
      checkDefinedAndNotNull(a);
    });

    it('should not throw for a defined value that is falsy', () => {
      checkDefinedAndNotNull(false);
      checkDefinedAndNotNull('');
      checkDefinedAndNotNull(0);
    });

    it('should throw for an undefined expressions', () => {
      expect(() => checkDefinedAndNotNull(undefined)).toThrow(PreconditionError);
    });

    it('should throw for an undefined expressions with custom message', () => {
      expect(() => checkDefinedAndNotNull(undefined, 'custom message')).toThrow(/custom message/);
    });

    it('should throw for an null expressions', () => {
      expect(() => checkDefinedAndNotNull(undefined)).toThrow(PreconditionError);
    });

    it('should throw for an null expressions with custom message', () => {
      expect(() => checkDefinedAndNotNull(undefined, 'my message')).toThrow(/my message/);
    });

    it('should narrow a type to non-nullable', () => {
      const maybeTwo = (): number | undefined => 2;
      expect(checkDefinedAndNotNull(maybeTwo()) + 1).toBe(3)
    });
  });
});
