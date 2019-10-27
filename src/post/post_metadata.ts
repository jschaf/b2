import * as dates from "../dates";
import * as strings from "../strings";
import Token from "markdown-it/lib/token";
import yaml from "js-yaml";

type Schema = Record<string, { type: 'string' | 'Date', isRequired: boolean }>;
const METADATA_SCHEMA: Schema = {
  slug: {type: 'string', isRequired: true},
  date: {type: 'Date', isRequired: true},
  publish_state: {type: 'string', isRequired: false},
};

/** The metadata for a post including title, date, draft status, and others. */
export class PostMetadata {

  private constructor(public readonly schema: Schema) {
  }

  static of(schema: any): PostMetadata {
    return new PostMetadata(checkMetadataSchema(schema));
  }

  /** Parses the post metadata from an array of markdown tokens. */
  static parseFromMarkdownTokens(tokens: Token[]): PostMetadata {
    const maxTokensToSearch = 20;
    const index = tokens.findIndex(
        t => t.type === 'fence' && t.content.startsWith('# Metadata'));
    if (index === -1) {
      throw new Error(`Unable to find a YAML metadata section in `
          + `the first ${maxTokensToSearch} tokens.`)
    }

    const token = tokens[index];
    const rawYaml = yaml.safeLoad(token.content);
    return PostMetadata.of(rawYaml);
  }
}

const checkMetadataSchema = (metadata: any): typeof METADATA_SCHEMA => {
  for (const [key, {isRequired}] of Object.entries(METADATA_SCHEMA)) {
    if (isRequired && !metadata.hasOwnProperty(key)) {
      throw new Error(`YAML metadata missing required key ${key}.`);
    }
  }

  for (const [key, value] of Object.entries(metadata)) {
    if (!METADATA_SCHEMA.hasOwnProperty(key)) {
      throw new Error((`Extra property key ${key} in YAML.`));
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
