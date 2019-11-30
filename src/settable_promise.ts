/* eslint-disable @typescript-eslint/no-explicit-any */
/**
 * A Promise whose result can be set from a value, another promise, or from
 * a rejection with a method call.  The promise can only be set once.
 */
export class SettablePromise<T> implements Promise<T> {
  // Override from Promise.
  readonly [Symbol.toStringTag]: string;

  private readonly hostPromise: Promise<T>;
  private resolve!: (value?: T | PromiseLike<T>) => void;
  private reject!: (reason?: any) => void;
  private wasSet: boolean = false;

  private constructor() {
    this.hostPromise = new Promise(
      (
        resolve: (value?: T | PromiseLike<T>) => void,
        reject: (reason?: any) => void
      ) => {
        this.resolve = resolve;
        this.reject = reject;
      }
    );
  }

  /**
   * Creates a new SettablePromise that can be resolved or reject by a later
   * method call.
   */
  static create<T>(): SettablePromise<T> {
    return new SettablePromise<T>();
  }

  private assertNotYetSet(): void {
    if (this.wasSet) {
      throw new Error(
        'Cannot set value of this SettablePromise because it was already set'
      );
    }
  }

  /**
   * Sets the result of this promise.
   *
   * Throws an error if the promise was already set.
   */
  set(value: T): void {
    this.setPromise(Promise.resolve(value));
  }

  /**
   * Sets the result of this promise to match the supplied promise once it
   * resolves or rejects.
   *
   * Throws an error if the promise was already set.
   */
  setPromise(promise: Promise<T>): void {
    this.assertNotYetSet();
    this.wasSet = true;
    promise.then(
      v => this.resolve(v),
      r => this.reject(r)
    );
  }

  /**
   * Sets the rejection of this promise.
   *
   * Throws an error if the promise was already set.
   */
  setReject(err?: any): void {
    this.assertNotYetSet();
    this.wasSet = true;
    this.reject(err);
  }

  then<TResult1 = T, TResult2 = never>(
    onfulfilled?:
      | ((value: T) => PromiseLike<TResult1> | TResult1)
      | undefined
      | null,
    onrejected?:
      | ((reason: any) => PromiseLike<TResult2> | TResult2)
      | undefined
      | null
  ): Promise<TResult1 | TResult2> {
    return this.hostPromise.then(onfulfilled, onrejected);
  }

  catch<TResult = never>(
    onrejected?:
      | ((reason: any) => PromiseLike<TResult> | TResult)
      | undefined
      | null
  ): Promise<T | TResult> {
    return this.hostPromise.catch(onrejected);
  }

  finally(onfinally?: (() => void) | undefined | null): Promise<T> {
    return this.hostPromise.finally(onfinally);
  }
}
