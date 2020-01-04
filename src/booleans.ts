export const isBoolean = (b: unknown): b is boolean => {
  if (typeof b === 'boolean') {
    return true;
  }
  // Check for Boolean constructor
  return (
    typeof b === 'object' && b !== null && typeof b.valueOf() === 'boolean'
  );
};

export const isOptionalBoolean = (b: unknown): b is boolean | undefined => {
  return b === undefined || isBoolean(b);
};
