import { SettablePromise } from '//settable_promise';

test('resolves when set is called', async () => {
  const p = SettablePromise.create<number>();

  p.set(1);

  const value = await p;
  expect(value).toBe(1);
});

test('resolves when setPromise is called', async () => {
  const p = SettablePromise.create<number>();

  p.setPromise(Promise.resolve(1));

  const value = await p;
  expect(value).toBe(1);
});

test('rejects when setPromise with rejected value is called', async () => {
  const p = SettablePromise.create();

  const errValue = 99;
  p.setPromise(Promise.reject(errValue));

  await expect(p).rejects.toEqual(errValue);
});

test('rejects when setReject is called', async () => {
  const p = SettablePromise.create();

  const errMessage = 'A message';
  p.setReject(errMessage);

  await expect(p).rejects.toEqual(errMessage);
});

test('errors when set is called twice', () => {
  const p = SettablePromise.create();

  p.set(1);

  expect(() => p.set(1)).toThrow();
});

test('errors when set is called followed by setPromise', async () => {
  const p = SettablePromise.create();

  p.set(1);

  expect(() => p.setPromise(Promise.resolve(2))).toThrow();
  await p;
});

test('errors when reject is called twice', async () => {
  const p = SettablePromise.create();

  p.setReject(1);

  expect(() => p.setReject(2)).toThrow();
  await expect(p).rejects.toEqual(1);
});

test('calls functions registered with `then`', async () => {
  const p: SettablePromise<number> = SettablePromise.create();
  const firstThen = p.then(x => x + 1);
  const secondThen = p.then(x => x + 2);

  p.set(10);

  expect(await firstThen).toBe(11);
  expect(await secondThen).toBe(12);
});

test('calls functions registered with `catch`', async () => {
  const p: SettablePromise<number> = SettablePromise.create();
  const firstCatch = p.catch(x => x + 1);
  const secondCatch = p.catch(x => x + 2);

  p.setReject(10);

  expect(await firstCatch).toBe(11);
  expect(await secondCatch).toBe(12);
});

test('calls functions registered with `finally`', async () => {
  const p: SettablePromise<number> = SettablePromise.create();
  let firstValue: number = -1;
  const firstFinally = p.finally(() => {
    firstValue = 1;
  });
  let secondValue: number = -1;
  const secondFinally = p.finally(() => {
    secondValue = 2;
  });

  p.set(10);

  expect(await firstFinally).toBe(10);
  expect(firstValue).toBe(1);
  expect(await secondFinally).toBe(10);
  expect(secondValue).toBe(2);
});
