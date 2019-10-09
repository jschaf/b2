/**
 * Returns true if value is a string.
 *
 * See https://stackoverflow.com/a/9436948/30900.
 */
export const isString = (value: any): boolean => {
  return typeof value === 'string' || value instanceof String;
};
