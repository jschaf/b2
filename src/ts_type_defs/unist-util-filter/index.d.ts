declare module 'unist-util-filter' {
  import * as unist from 'unist';
  import * as unistIs from 'unist-util-is';

  interface FilterFn {
    (
      tree: unist.Node,
      options?: { cascade: boolean },
      test?: unistIs.Test<unist.Node>
    ): boolean;

    (tree: unist.Node, test: unistIs.Test<unist.Node>): boolean;
  }

  const filter: FilterFn;
  export = filter;
}
