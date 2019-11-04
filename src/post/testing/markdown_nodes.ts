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

export const stripPositions = (node: PostNode): PostNode => {
  const forceDelete = true;
  return new PostNode(node.metadata, removePosition(node.node, forceDelete));
};
