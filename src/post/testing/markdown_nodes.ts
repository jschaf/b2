import {removePositionInfo} from '//unist/nodes';
import * as toml from '@iarna/toml';
import { PostNode } from '../post_parser';

interface MdNode {
  type: string;
  children?: MdNode[];

  [key: string]: unknown;
}

const mdNode = (
  type: string,
  params: Record<string, unknown>,
  children?: MdNode[]
): MdNode => {
  const childObj = children == null ? {} : { children };
  return { type, ...params, ...childObj };
};

export const mdRoot = (children: MdNode[]): MdNode => {
  return mdNode('root', {}, children);
};

export const mdHeading = (depth: number, children: MdNode[]): MdNode => {
  return mdNode('heading', { depth }, children);
};

export const mdHeading1 = (child: string): MdNode => {
  return mdHeading(1, [mdText(child)]);
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

export const mdFrontmatterToml = (value: toml.JsonMap): MdNode => {
  let raw = toml
    .stringify(value)
    .trimEnd()
    .replace(/T00:00:00.000Z/, '');
  return {
    type: 'toml',
    value: raw,
  };
};

export const mdCode = (value: string): MdNode => {
  return mdNode('code', { value });
};

export const stripPositions = (node: PostNode): PostNode => {
  removePositionInfo(node.node);
  return node;
};
