import { checkState } from '//asserts';
import { Unzipper } from '//zip_files';

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
