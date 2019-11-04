import {checkState} from '../asserts';
import {Unzipper} from '../zip_files';

/**
 * An append-only, in-memory representation of a post.
 */
export class Mempost {
  static MD_CONTENT_PATH = 'content.md';

  private readonly entriesByPath = new Map<string, Buffer>();

  private constructor() {}

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
    checkState(!this.entriesByPath.has(path),
        `Expected no existing entry for path: '${path}'`);
    this.entriesByPath.set(path, contents);
  }

  addUtf8Entry(path: string, contents: string): void {
    this.addEntry(path, Buffer.from(contents, 'utf8'));
  }

  getEntry(path: string): Buffer | undefined {
    const buf = this.entriesByPath.get(path);
    return buf === undefined ? undefined : buf;
  }

  getUtf8Entry(path: string): string | undefined {
    const buf = this.getEntry(path);
    return buf === undefined ? undefined : buf.toString('utf8');
  }

  entries(): IterableIterator<[string, Buffer]> {
    return this.entriesByPath[Symbol.iterator]();
  }
}

