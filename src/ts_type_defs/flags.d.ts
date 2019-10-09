declare module 'flags' {

  interface Flag<T> {
    name: string;
    defaultValue: T;
    currentValue: T;
    isSet: boolean;

    setDefault(defaultValue: T): Flag<T>;
    setDescription(description: string): Flag<T>;
    setValidator(validator: (input: string) => void): Flag<T>;
    setSecret(secret: boolean): Flag<T>;
  }
  const parse: (argv?: String[]) => void;
  const reset: () => void;
  const get: (name: string) => Flag<unknown>;

  const defineString: (name: string) => Flag<string>;
  const defineInteger: (name: string) => Flag<number>;
  const defineNumber: (name: string) => Flag<number>;
  const defineStringList: (name: string) => Flag<String[]>;
  const defineMultiString: (name: string) => Flag<number>;

}
