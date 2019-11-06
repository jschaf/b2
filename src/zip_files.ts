import * as unzip from 'yauzl';
import * as yazl from 'yazl';
import * as streams from './streams';
import { SettablePromise } from '//settable_promise';
import { checkArg, checkDefinedAndNotNull } from '//asserts';

/** An entry from a zip file. */
export class ZipFileEntry {
  private constructor(readonly filePath: string, readonly contents: Buffer) {
    checkArg(
      filePath === '' || filePath.charAt(0) !== '/',
      `The file path for zip file cannot be absolute but had '${filePath}'.`
    );
  }

  static ofBuffer(filePath: string, contents: Buffer): ZipFileEntry {
    return new ZipFileEntry(filePath, contents);
  }

  static ofUtf8(filePath: string, contents: string): ZipFileEntry {
    return new ZipFileEntry(filePath, Buffer.from(contents, 'utf8'));
  }
}

export class Zipper {
  static async zip(entries: ZipFileEntry[]): Promise<Buffer> {
    const zipFile = new yazl.ZipFile();
    for (const { filePath, contents } of entries) {
      zipFile.addBuffer(contents, filePath);
    }
    zipFile.end();

    const bufs: Buffer[] = [];
    for await (const chunk of zipFile.outputStream) {
      bufs.push(chunk as Buffer);
    }
    return Buffer.concat(bufs);
  }
}

export class Unzipper {
  static async unzip(buf: Buffer): Promise<ZipFileEntry[]> {
    const zipFile = await unzipFromBuffer(buf);
    return readAllEntries(zipFile);
  }
}

/**
 * Unzips a buffer into a zip file.
 *
 * The returned zip file is lazily parsed and won't have any data until
 * zipFile.readEntry() is called.
 */
const unzipFromBuffer = (buf: Buffer): Promise<unzip.ZipFile> => {
  const promisedResult = SettablePromise.create<unzip.ZipFile>();
  unzip.fromBuffer(buf, { lazyEntries: true }, (err, zipFile) => {
    if (err) {
      promisedResult.setReject(err);
      return;
    }
    promisedResult.set(checkDefinedAndNotNull(zipFile));
  });
  return promisedResult;
};

/**
 * Reads a lazily loaded ZipFile into list of file entries. Each entry
 * consists of a file path and the file contents as a Buffer.
 */
const readAllEntries = async (
  zipFile: unzip.ZipFile
): Promise<ZipFileEntry[]> => {
  const promisedResults = SettablePromise.create<ZipFileEntry[]>();
  const results: Promise<ZipFileEntry>[] = [];

  // Directory file names end with '/'. Entries for directories
  // themselves are optional. An entry's fileName implicitly
  // requires its parent directories to exist.
  const isDir = (f: string): boolean => /\/$/.test(f);

  const readFileEntry = (entry: unzip.Entry): Promise<ZipFileEntry> => {
    const fileEntry = SettablePromise.create<ZipFileEntry>();
    zipFile.openReadStream(entry, async (err, readStream) => {
      if (err) {
        return fileEntry.setReject(err);
      }
      const contents = await streams.toBuffer(
        checkDefinedAndNotNull(readStream)
      );
      return fileEntry.set({ filePath: entry.fileName, contents });
    });
    return fileEntry;
  };

  zipFile.readEntry();
  zipFile.on('entry', async (entry: unzip.Entry) => {
    if (isDir(entry.fileName)) {
      zipFile.readEntry();
    } else {
      results.push(readFileEntry(entry));
      zipFile.readEntry();
    }
  });
  zipFile.once('end', () => {
    return promisedResults.setPromise(Promise.all(results));
  });

  return promisedResults;
};
