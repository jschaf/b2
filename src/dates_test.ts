import * as dates from "./dates";
import {isValidDate} from "./dates";

test('isValidDate works', () => {
  expect(isValidDate(0)).toBe(false);
  expect(isValidDate(null)).toBe(false);
  expect(isValidDate('foo')).toBe(false);

  expect(isValidDate(new Date('2019-10-08'))).toBe(true);
});

test('fromISO works', () => {
  expect(dates.fromISO('2019-10-08')).toEqual(new Date('2019-10-08'));
});
