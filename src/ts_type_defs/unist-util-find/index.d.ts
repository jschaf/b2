declare module 'unist-util-find' {
  import * as unist from 'unist';

  const find: <T extends unist.Node>(
    tree: unist.Node,
    test: (node: unist.Node, index?: number, parent?: unist.Node) => node is T
  ) => T | undefined;
  export = find;
}
