import * as dates from '//dates';
import * as md from '//post/mdast/nodes';
import * as strings from '//strings';
import * as unistNodes from '//unist/nodes';

import * as toml from '@iarna/toml';
import yaml from 'js-yaml';
import * as unist from 'unist';

type Schema = Record<string, { type: 'string' | 'Date'; isRequired: boolean }>;
const METADATA_SCHEMA: Schema = {
  slug: { type: 'string', isRequired: true },
  date: { type: 'Date', isRequired: true },
  publishState: { type: 'string', isRequired: false },
};

type Metadata = {
  slug: string;
  date: Date;
  publishState?: string;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
} & Record<string, any>;

/** The metadata for a post including title, date, draft status, and others. */
export class PostMetadata {
  private constructor(
    public readonly slug: string,
    public readonly date: Date,
    public readonly schema: Schema
  ) {}

  static empty(): PostMetadata {
    return PostMetadata.parse({ slug: '', date: dates.fromISO('1970-01-01') });
  }

  static parse(schema: Record<string, unknown>): PostMetadata {
    checkValidSchema(schema);
    return new PostMetadata(schema.slug, schema.date, schema);
  }

  static isCodeMetadataNode = (
    n: unist.Node
  ): n is { type: 'code'; value: string } =>
    md.isCode(n) && n.value.startsWith('# Metadata');

  /** Parses the post metadata from an mdast node. */
  static parseFromMdast(tree: unist.Node): PostMetadata | null {
    const t = this.extractFromTomlFrontmatter(tree);
    if (t && isValidSchema(t)) {
      return PostMetadata.parse(t);
    }

    const m = this.extractFromMetadataCodeNode(tree);
    if (m && isValidSchema(m)) {
      return PostMetadata.parse(m);
    }
    return null;
  }

  private static extractFromTomlFrontmatter(
    tree: unist.Node
  ): Record<string, unknown> | null {
    const node = unistNodes.findNode(tree, md.isToml);
    if (node === null) {
      return null;
    }
    return toml.parse(node.value);
  }

  private static extractFromMetadataCodeNode(
    tree: unist.Node
  ): Record<string, unknown> | null {
    const node = unistNodes.findNode(tree, PostMetadata.isCodeMetadataNode);
    if (node === null) {
      return null;
    }
    return yaml.safeLoad(node.value);
  }

  /**
   * Normalizes an mdast tree by ensuring the metadata node is toml and it's the
   * first child in the tree.
   */
  static normalizeMdast(tree: unist.Node): unist.Node {
    if (!md.isParent(tree)) {
      return tree;
    }

    const tomlData = unistNodes.findNode(tree, md.isToml);
    const codeData = unistNodes.findNode(tree, this.isCodeMetadataNode);

    if (tomlData !== null) {
      // Move toml to the first node in mdast.
      unistNodes.removeNode(tree, md.isToml);
      tree.children.unshift(tomlData);

      if (codeData !== null) {
        // Remove the code metadata and assume toml is canonical.
        unistNodes.removeNode(tree, this.isCodeMetadataNode);
        return tree;
      } else {
        // Nothing to do because only toml node exists.
        return tree;
      }
    } else {
      if (codeData !== null) {
        // Convert code metadata into toml and remove code metadata.
        unistNodes.removeNode(tree, this.isCodeMetadataNode);
        const schema = yaml.safeLoad(codeData.value);
        const t = md.tomlFrontmatter(schema);
        tree.children.unshift(t);
        return tree;
      } else {
        // No metadata so nothing to do.
        return tree;
      }
    }
  }
}

const isValidSchema = (m: Record<string, unknown>): m is Metadata => {
  try {
    checkValidSchema(m);
    return true;
  } catch (e) {
    return false;
  }
};

// eslint-disable-next-line @typescript-eslint/no-explicit-any
function checkValidSchema(m: Record<string, unknown>): asserts m is Metadata {
  for (const [key, { isRequired }] of Object.entries(METADATA_SCHEMA)) {
    if (isRequired && !m.hasOwnProperty(key)) {
      throw new Error(`Metadata missing required key: '${key}'.`);
    }
  }

  for (const [key, value] of Object.entries(m)) {
    if (!METADATA_SCHEMA.hasOwnProperty(key)) {
      throw new Error(
        `Extra property key '${key}' in YAML. ` +
          `Expected only keys: ${Object.keys(METADATA_SCHEMA).join(', ')}`
      );
    }
    const schemaDef = METADATA_SCHEMA[key];
    switch (schemaDef.type) {
      case 'Date':
        if (!dates.isValidDate(value)) {
          throw new Error(`Invalid date: ${value} for key: ${key}.`);
        }
        break;
      case 'string':
        if (!strings.isString(value)) {
          throw new Error(`Expected string for key ${key} but got ${value}.`);
        }
        break;
    }
  }
}
