import { isString } from '//strings';

export const newTypeGuardCheck = <T extends string, E extends string>(
  enumVariable: { [key in T]: E }
): ((v: unknown) => v is E) => {
  const enumValues = Object.values(enumVariable);
  return (value: unknown): value is E => {
    return isString(value) && enumValues.includes(value);
  };
};
