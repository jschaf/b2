import { checkState } from '//asserts';
import { isValidDate } from '//dates';
import * as dates from '//dates';
import * as md from '//post/mdast/nodes';
import { isString } from '//strings';
import * as unistNodes from '//unist/nodes';
import * as enums from '//enums';

import * as toml from '@iarna/toml';
import yaml from 'js-yaml';
import * as unist from 'unist';

export enum PostType {
  Post = 'post',
  LandingPage = 'landing_page',
}

const isPostType = enums.newTypeGuardCheck(PostType);

export enum PublishState {
  Draft = 'draft',
  Published = 'published',
}

const isPublishState = enums.newTypeGuardCheck(PublishState);

type Metadata = {
  slug: string;
  date: Date;
  publishState: PublishState;
  postType: PostType;
};

/** The metadata for a post including title, date, draft status, and others. */
export class PostMetadata {
  private constructor(
    public readonly slug: string,
    public readonly date: Date,
    public readonly postType: PostType,
    public readonly schema: Record<string, unknown>
  ) {}

  static empty(): PostMetadata {
    return PostMetadata.parse({ slug: '', date: dates.fromISO('1970-01-01') });
  }

  static parse(schema: Record<string, unknown>): PostMetadata {
    const m = extractValidMetadata(schema);
    return new PostMetadata(m.slug, m.date, m.postType, schema);
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

const extractValidMetadata = (m: Record<string, unknown>): Metadata => {
  const date = m.date;
  checkState(isValidDate(date), `date must be valid but had ${date}`);

  const publishState = (m.publish_state || PublishState.Draft) as PublishState;
  checkState(isPublishState(publishState));

  const postType = m.post_type || PostType.Post;
  checkState(isPostType(postType), `post_type is not valid: ${postType}`);

  const slug = m.slug || '';
  checkState(isString(slug), `slug must be a string but had ${slug}`);

  return { date, publishState, postType, slug };
};

const isValidSchema = (m: Record<string, unknown>): m is Metadata => {
  try {
    extractValidMetadata(m);
    return true;
  } catch (e) {
    return false;
  }
};
