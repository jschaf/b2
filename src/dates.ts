/**
 * Returns true if a date is valid.
 *
 * See https://stackoverflow.com/questions/643782
 */
export const isValidDate = (date: any): boolean => {
  return !!date
      && Object.prototype.toString.call(date) === "[object Date]"
      // @ts-ignore - check invalid dates
      && !isNaN(date);
};
