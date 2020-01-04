export const isObject = (o: unknown): o is Record<string, unknown> => {
  return typeof o === 'object' && o !== null;
};

export const isOptionalObject = (
  o: unknown
): o is Record<string, unknown> | undefined => {
  return isObject(o) || o === undefined;
};

export const lossyClone = <T extends object>(o: T): T => {
  return JSON.parse(JSON.stringify(o));
};

export const isEmpty = (o: object): boolean => {
  // https://stackoverflow.com/a/32108184/30900
  return Object.entries(o).length === 0 && o.constructor === Object;
};
