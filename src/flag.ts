type RawFlags = Map<string, string>;

export class FlagVal<T> {
  private constructor(
      readonly name: string,
      readonly defaultValue: T,
      public currentValue: T,
      public isSet: boolean,
  ) {
  }

  static create<U>(
      name: string,
      defaultValue: U,
      currentValue: U,
      isSet: boolean,
  ): FlagVal<U> {
    return new FlagVal(name, defaultValue, currentValue, isSet);
  }
}

export class FlagSet {
  private constructor() {
  }

  static create(): FlagSet {
    return new FlagSet();
  }

  static newInt(name: string, value: number, usage: string): FlagVal<number> {
    return FlagVal.create(name, value, value, )


  }

}
