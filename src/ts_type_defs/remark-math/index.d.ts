declare module 'remark-math' {
  import { Plugin } from 'unified';

  interface RemarkMath extends Plugin<[Partial<RemarkMathOptions>?]> {}

  type RemarkMathOptions = {};

  const remarkMath: RemarkMath;
  export = remarkMath;
}
