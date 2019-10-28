declare module 'unist-util-remove' {
  import * as unist from 'unist';

  const remove: <T extends unist.Node>(
    tree: unist.Node,
    test: (node: unist.Node, index?: number, parent?: unist.Node) => node is T
  ) => unist.Node | null;
  export = remove;
}
