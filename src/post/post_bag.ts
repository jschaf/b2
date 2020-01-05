import { checkState } from '//asserts';
import { Unzipper, ZipFileEntry } from '//zip_files';
import { PostNode, PostParser, TEXT_PACK_BUNDLE_PREFIX } from '//post/parser';

export class PostBag {
  private constructor(readonly postNode: PostNode) {}

  static fromMarkdown(mainText: string): PostBag {
    return new PostBag(PostParser.create().parseMarkdown(mainText));
  }

  static async fromTextPack(textPack: Buffer): Promise<PostBag> {
    const entries = await Unzipper.unzip(textPack);
    const mainText = findMainText(entries);
    return PostBag.fromMarkdown(mainText);
  }
}

const findMainText = (entries: ZipFileEntry[]): string => {
  const texts = entries.filter(
    e => e.filePath === TEXT_PACK_BUNDLE_PREFIX + '/text.md'
  );
  checkState(texts.length === 1, 'Expected a single text.md file in TextPack.');
  return texts[0].contents.toString('utf8');
};
