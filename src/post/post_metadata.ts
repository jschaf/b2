import {findNode} from '//unist/nodes';
import yaml from 'js-yaml';
import * as unist from 'unist';
import { checkDefined } from '//asserts';
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

  static of(schema: unknown): PostMetadata {
    const validated = checkMetadataSchema(schema);
    return new PostMetadata(validated['slug'], validated['date'], validated);
  }

  static isMetadataNode = (
    n: unist.Node
  ): n is { type: 'code'; value: string } =>
    n.type === 'code' && isString(n.value) && n.value.startsWith('# Metadata');

  /** Parses the post metadata from an array of markdown tokens. */
  static parseFromMarkdownAST(tree: unist.Node): PostMetadata {
    const node = checkDefined(
      findNode(tree, PostMetadata.isMetadataNode),
      "No nodes found that match YAML code block beginning with '# Metadata'."
    );
    const rawYaml = yaml.safeLoad(node.value);
    return PostMetadata.of(rawYaml);
  }

  static isTomlFrontmatterNode = (
    n: unist.Node
  ): n is { type: 'toml'; value: string } => {
    return n.type === 'toml' && isString(n.value);
  };

  static parseFromTomlFrontmatter(tree: unist.Node): PostMetadata {
    const node = checkDefined(
      findNode(tree, PostMetadata.isTomlFrontmatterNode),
      'No nodes found that match a TOML frontmatter block.'
    );
    const rawToml = toml.parse(node.value);
    return PostMetadata.of(rawToml);
  }
}

// eslint-disable-next-line @typescript-eslint/no-explicit-any
const checkMetadataSchema = (metadata: any): Metadata => {
  for (const [key, { isRequired }] of Object.entries(METADATA_SCHEMA)) {
    if (isRequired && !metadata.hasOwnProperty(key)) {
      throw new Error(`Metadata missing required key: '${key}'.`);
    }
  }

  for (const [key, value] of Object.entries(metadata)) {
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
  return metadata;
};
