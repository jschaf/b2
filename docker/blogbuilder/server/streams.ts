import * as stream from "stream";

/** Collects each chunk of a stream into a separate index in an array. */
export const collectToArray = <T>(data: stream.Stream): Promise<T[]> => {
  const chunks: T[] = [];
  return new Promise((resolve, reject) => {
    data.on('data', chunk => chunks.push(chunk));
    data.on('error', reject);
    data.on('end', () => resolve(chunks))
  })
};

/** Creates a readable stream from an array. */
export const createFromArray = <T>(chunks: T[]): stream.Stream => {
  const s = new stream.Readable({objectMode: true});
  for (const c of chunks) {
    s.push(c);
  }
  // End of stream.
  s.push(null);
  return s;
};

/** Creates a UTF-8 string from a stream of bytes (Uint8Array). */
export const toUtf8String = (data: stream.Stream): Promise<string> => {
  const chunks: Uint8Array[] = [];
  return new Promise((resolve, reject) => {
    data.on('data', chunk => chunks.push(chunk));
    data.on('error', reject);
    data.on('end', () => resolve(Buffer.concat(chunks).toString('utf8')))
  })
};

