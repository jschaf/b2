import removePosition from 'unist-util-remove-position';
import { PostNode } from '../post_parser';

type MdNode = { type: string; children?: MdNode[] };

const mdNode = (type: string, params: Object, children?: MdNode[]): MdNode => {
  const childObj = children == null ? {} : { children };
  return { type, ...params, ...childObj };
};

export const mdRoot = (children: MdNode[]): MdNode => {
  return mdNode('root', {}, children);
};

export const mdHeading = (depth: number, children: MdNode[]): MdNode => {
  return mdNode('heading', { depth }, children);
};

export const mdText = (value: string): MdNode => {
  return mdNode('text', { value });
};

export const mdPara = (children: MdNode[]): MdNode => {
  return mdNode('paragraph', {}, children);
};

export const mdParaText = (value: string): MdNode => {
  return mdPara([mdText(value)]);
};

export const mdOrderedList = (children: MdNode[]): MdNode => {
  return mdNode(
    'list',
    {
      ordered: true,
      spread: false,
      start: 1,
      type: 'list',
    },
    children.map(c => (c.type === 'listItem' ? c : mdListItem([c])))
  );
};

export const mdListItem = (children: MdNode[]): MdNode => {
  return mdNode('listItem', { spread: false, checked: null }, children);
};

export const stripPositions = (node: PostNode): PostNode => {
  const forceDelete = true;
  return new PostNode(node.metadata, removePosition(node.node, forceDelete));
};
