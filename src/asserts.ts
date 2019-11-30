/** The error thrown when checkArg fails. */
export class PreconditionError extends Error {
  constructor(message?: string) {
    super(message);
  }
}

/**
 * Ensures the truth of an expression involving one or more parameters to the
 * calling method.
 *
 * Example:
 *
 * function sqrt(x: number): number {
 *     checkArg(x >= 0, `negative value: ${x}`);
 *     // calculate the square root.
 * }
 *
 * Throws PreconditionError if the expression is false.
 */
export function checkArg(expression: boolean, errorMsg?: string): asserts expression {
  if (!expression) {
    throw new PreconditionError(
        errorMsg || 'Expression evaluated to false but expected it to be true.'
    );
  }
}

/**
 * Ensures the truth of an expression involving the state of the calling instance,
 * but not involving any parameters to the calling method.
 *
 * For example, a method might verify that a row from the database has a
 * specific value.
 *
 * Throws PreconditionError if the expression is false.
 */
export function checkState(expression: boolean, errorMsg?: string): asserts expression {
  if (!expression) {
    throw new PreconditionError(
        errorMsg || 'Expression evaluated to false but expected it to be true.'
    );
  }
}

/** Ensures an expression is both defined and not null. */
export const checkDefined = <T>(
  expression: T | undefined | null,
  errorMsg?: string
): NonNullable<T> => {
  if (expression === undefined) {
    throw new PreconditionError(
      errorMsg || 'Expression was undefined but expected a defined expression.'
    );
  }
  if (expression === null) {
    throw new PreconditionError(
      errorMsg || 'Expression was null but expected a non-null expression.'
    );
  }
  return expression as NonNullable<T>;
};
