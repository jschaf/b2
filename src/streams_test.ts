import * as streams from '//streams';

test('createFromArray creates a stream', async () => {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const roundTrip = async (arr: any[]): Promise<any> =>
    streams.collectToArray(streams.createFromArray(arr));

  expect(await roundTrip([])).toStrictEqual([]);
  expect(await roundTrip([1])).toStrictEqual([1]);
  expect(await roundTrip([1, 0, -1, 3])).toStrictEqual([1, 0, -1, 3]);
  expect(await roundTrip([1, 'foo', -1, 3])).toStrictEqual([1, 'foo', -1, 3]);
});

test('toUtf8String creates a string', async () => {
  const roundTrip = async (str: string): Promise<string> => {
    const codePoints = Array.from(str).map(c => Uint8Array.of(c.charCodeAt(0)));
    return streams.toUtf8String(streams.createFromArray(codePoints));
  };

  expect(await roundTrip('')).toBe('');
  expect(await roundTrip('f')).toBe('f');
  expect(await roundTrip('foo')).toBe('foo');
});
