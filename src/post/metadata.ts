import * as unistNodes from '//unist/nodes';
import yaml from 'js-yaml';
import * as unist from 'unist';
import * as md from '//post/mdast/nodes';
import * as dates from '//dates';
import * as strings from '//strings';
import { isString } from '//strings';

import * as toml from '@iarna/toml';
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

  static isMetadataNode = (
    n: unist.Node
  ): n is { type: 'code'; value: string } =>
    n.type === 'code' && isString(n.value) && n.value.startsWith('# Metadata');

  /** Parses the post metadata from an mdast node. */
  static parseFromMdast(tree: unist.Node): PostMetadata | null {
    const t = this.extractFromTomlFrontmatter(tree);
    if (t && isValidSchema(t)) {
      return PostMetadata.parse(t);
    }

    const m = this.extractFromMetadataCodeBlock(tree);
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

  private static extractFromMetadataCodeBlock(
    tree: unist.Node
  ): Record<string, unknown> | null {
    const node = unistNodes.findNode(tree, PostMetadata.isMetadataNode);
    if (node === null) {
      return null;
    }
    return yaml.safeLoad(node.value);
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
