export const isObject = (o: unknown): o is Record<string, unknown> => {
  return typeof o === 'object' && o !== null;
};

export const lossyClone = <T extends object>(o: T): T => {
  return JSON.parse(JSON.stringify(o));
};
