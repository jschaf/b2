export const isNumber = (n: unknown): n is number => {
  return typeof n === 'number';
};

export const isOptionalNumber = (n: unknown): n is number | undefined => {
  return n === undefined || isNumber(n);
};
