import { checkState } from '//asserts';
import { Unzipper } from '//zip_files';
import rehypeFormat from 'rehype-format';
import rehypeParse from 'rehype-parse';
import rehypeStringify from 'rehype-stringify';
import unified from 'unified';

/**
 * An append-only, in-memory representation of a post.
 */
export class Mempost {
  private readonly entriesByPath = new Map<string, string | Buffer>();

  private constructor() {}

  static create(): Mempost {
    return new Mempost();
  }

  static ofUtf8Entry(path: string, contents: string): Mempost {
    const m = Mempost.create();
    m.addEntry(path, contents);
    return m;
  }

  static async fromTextPack(textPack: Buffer): Promise<Mempost> {
    const entries = await Unzipper.unzip(textPack);
    const mp = Mempost.create();
    for (const entry of entries) {
      mp.addEntry(entry.filePath, entry.contents);
    }
    return mp;
  }

  addEntry(path: string, contents: string | Buffer): void {
    checkState(
      !this.entriesByPath.has(path),
      `Expected no existing entry for path: '${path}'`
    );
    this.entriesByPath.set(path, contents);
  }

  getEntry(path: string): string | Buffer | undefined {
    return this.entriesByPath.get(path);
  }

  toRecord(): Record<string, string | Buffer> {
    const results: Record<string, string | Buffer> = {};
    for (const [path, content] of this.entriesByPath) {
      results[path] = content;
    }
    return results;
  }
}

/**
 * Converts a Buffer to a UTF-8 string if possible. Otherwise, return the buffer.
 *
 * Intended purposed is to produce cleaner error messages.
 */
export const normalizeMempostEntry = (
  path: string,
  buf: string | Buffer
): string => {
  try {
    if (path.endsWith('.html')) {
      return normalizeHTML(buf);
    }
    return buf.toString('utf8');
  } catch (e) {
    return buf.toString('utf8');
  }
};

const htmlProcessor = unified()
  .use(rehypeParse)
  .use(rehypeFormat)
  .use(rehypeStringify);

export const normalizeHTML = (contents: string | Buffer): string => {
  const vFile = htmlProcessor.processSync(contents);
  return vFile.contents.toString('utf8');
};
