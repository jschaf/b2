declare module 'unist-util-remove-position' {
  import * as unist from 'unist';
  const removePosition: (node: unist.Node, force?: boolean) => unist.Node;
  export = removePosition;
}
