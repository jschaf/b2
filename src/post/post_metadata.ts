import * as dates from '../dates';
import * as unist from 'unist';
import * as strings from '../strings';
import { isString } from '../strings';
import findNode from 'unist-util-find';
import yaml from 'js-yaml';

type Schema = Record<string, { type: 'string' | 'Date'; isRequired: boolean }>;
const METADATA_SCHEMA: Schema = {
  slug: { type: 'string', isRequired: true },
  date: { type: 'Date', isRequired: true },
  publish_state: { type: 'string', isRequired: false },
};

/** The metadata for a post including title, date, draft status, and others. */
export class PostMetadata {
  private constructor(public readonly schema: Schema) {}

  static of(schema: any): PostMetadata {
    return new PostMetadata(checkMetadataSchema(schema));
  }

  static isMetadataNode(n: unist.Node): n is { type: 'code'; value: string } {
    return (
      n.type === 'code' && isString(n.value) && n.value.startsWith('# Metadata')
    );
  }

  /** Parses the post metadata from an array of markdown tokens. */
  static parseFromMarkdownAST(tree: unist.Node): PostMetadata {
    const node = findNode(tree, PostMetadata.isMetadataNode);
    if (node === undefined) {
      throw new Error(
        "No nodes found that match YAML code block beginning with '# Metadata'."
      );
    }
    const rawYaml = yaml.safeLoad(node.value);
    return PostMetadata.of(rawYaml);
  }
}

const checkMetadataSchema = (metadata: any): typeof METADATA_SCHEMA => {
  for (const [key, { isRequired }] of Object.entries(METADATA_SCHEMA)) {
    if (isRequired && !metadata.hasOwnProperty(key)) {
      throw new Error(`YAML metadata missing required key ${key}.`);
    }
  }

  for (const [key, value] of Object.entries(metadata)) {
    if (!METADATA_SCHEMA.hasOwnProperty(key)) {
      throw new Error(`Extra property key ${key} in YAML.`);
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
