import { Mempost } from '//post/mempost';

declare global {
  namespace jest {
    // We're merging jest namespace with our own custom matchers.
    interface Matchers<R, T> {
      toEqualMempost(
        value: Mempost | Record<string, string>
      ): CustomMatcherResult;
      toEqualHTML(value: string): CustomMatcherResult;
    }
  }
}
