import {checkState} from '//asserts';
import {Unzipper} from '//zip_files';
import rehypeFormat from 'rehype-format';
import rehypeParse from 'rehype-parse';
import rehypeStringify from 'rehype-stringify';
import unified from 'unified';

/**
 * An append-only, in-memory representation of a post.
 */
export class Mempost {
  private readonly entriesByPath = new Map<string, Buffer>();
  private readonly utf8EntriesByPath = new Map<string, string>();

  private constructor() {
  }

  static create(): Mempost {
    return new Mempost();
  }

  static ofUtf8Entry(path: string, contents: string): Mempost {
    const m = Mempost.create();
    m.addUtf8Entry(path, contents);
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

  addEntry(path: string, contents: Buffer): void {
    this.assertNotYetAdded(path);
    this.entriesByPath.set(path, contents);
  }

  addUtf8Entry(path: string, contents: string): void {
    this.assertNotYetAdded(path);
    this.utf8EntriesByPath.set(path, contents);
  }

  getEntry(path: string): Buffer | undefined {
    return this.entriesByPath.get(path);
  }

  getUtf8Entry(path: string): string | undefined {
    return this.utf8EntriesByPath.get(path);
  }

  *entries(): IterableIterator<[string, string | Buffer]> {
    for (const entry of this.entriesByPath) {
      yield entry;
    }
    for (const entry of this.utf8EntriesByPath) {
      yield entry;
    }
  }

  private assertNotYetAdded(path: string) {
    checkState(
        !this.entriesByPath.has(path) && !this.utf8EntriesByPath.has(path),
        `Expected no existing entry for path: '${path}'`
    );
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
