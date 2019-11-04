import * as stream from 'stream';

/** Collects each chunk of a stream into a separate index in an array. */
export const collectToArray = async <T>(
  data: stream.Readable
): Promise<T[]> => {
  const chunks: T[] = [];
  for await (const chunk of data) {
    chunks.push(chunk);
  }
  return chunks;
};

/** Creates a readable stream from an array. */
export const createFromArray = <T>(chunks: T[]): stream.Readable => {
  const s = new stream.Readable({ objectMode: true });
  for (const c of chunks) {
    s.push(c);
  }
  // End of stream.
  s.push(null);
  return s;
};

/** Creates a Buffer from a stream of bytes (Uint8Array). */
export const toBuffer = async (data: stream.Readable): Promise<Buffer> => {
  const chunks: Uint8Array[] = [];
  for await (const chunk of data) {
    chunks.push(chunk);
  }
  return Buffer.concat(chunks);
};

/** Creates a UTF-8 string from a stream of bytes (Uint8Array). */
export const toUtf8String = async (data: stream.Readable): Promise<string> => {
  const buf = await toBuffer(data);
  return buf.toString('utf8');
};
