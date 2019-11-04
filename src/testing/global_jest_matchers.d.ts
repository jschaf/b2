import {Mempost} from '../post/mempost';

declare global {
  namespace jest {
    // We're merging jest namespace with our own custom matchers.
    //@ts-ignore
    interface Matchers<R> {
      toEqualMempost(value: Mempost | Record<string, string>): CustomMatcherResult;
    }
  }
}
