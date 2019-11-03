import { Unzipper, ZipFileEntry, Zipper } from './zip_files';

describe('Zipper', () => {
  it('should roundtrip with 0 entries', async () => {
    const buf = await Zipper.zip([]);
    expect(await Unzipper.unzip(buf)).toEqual([]);
  });

  it('should roundtrip with 1 entry', async () => {
    const entry = ZipFileEntry.ofUtf8('some/path', 'foo');
    const buf = await Zipper.zip([entry]);
    expect(await Unzipper.unzip(buf)).toEqual([entry]);
  });

  it('should roundtrip with 2 entries', async () => {
    const entry1 = ZipFileEntry.ofUtf8('some/path', 'foo');
    const entry2 = ZipFileEntry.ofUtf8('some/path', 'bar');
    const buf = await Zipper.zip([entry1, entry2]);
    expect(await Unzipper.unzip(buf)).toEqual([entry1, entry2]);
  });
});

describe('ZipFileEntry', () => {
  it('should throw for an absolute path', () => {
    expect(() => ZipFileEntry.ofUtf8('/abs', '')).toThrow('/abs');
  });
});
