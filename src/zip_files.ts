import * as unzip from "yauzl";
import * as streams from './streams';
import {SettablePromise} from "./settable_promise";

// A single entry within a ZipFile.
export type FileEntry = { filePath: string, contents: Buffer };

/**
 * Unzips a buffer into a zip file.
 *
 * The returned zip file is lazily parsed and won't have any data until
 * zipFile.readEntry() is called.
 */
export const unzipFromBuffer = (buf: Buffer): Promise<unzip.ZipFile> => {
  const promisedResult = SettablePromise.create<unzip.ZipFile>();
  unzip.fromBuffer(buf, {lazyEntries: true}, (err, zipFile) => {
    if (err) {
      promisedResult.setReject(err);
      return;
    }
    promisedResult.set(zipFile!);
  });
  return promisedResult;
};

/**
 * Reads a lazily loaded ZipFile into list of file entries. Each entry
 * consists of a file path and the file contents as a Buffer.
 */
export const readAllEntries = async (zipFile: unzip.ZipFile): Promise<FileEntry[]> => {
  const promisedResults = SettablePromise.create<FileEntry[]>();
  const results: Promise<FileEntry>[] = [];

  // Directory file names end with '/'. Entries for directories
  // themselves are optional. An entry's fileName implicitly
  // requires its parent directories to exist.
  const isDir = (f: string) => /\/$/.test(f);

  const readFileEntry = (entry: unzip.Entry): Promise<FileEntry> => {
    const fileEntry = SettablePromise.create<FileEntry>();
    zipFile.openReadStream(entry, async (err, readStream) => {
      if (err) {
        return fileEntry.setReject(err);
      }
      const contents = await streams.toBuffer(readStream!);
      return fileEntry.set({filePath: entry.fileName, contents});
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
