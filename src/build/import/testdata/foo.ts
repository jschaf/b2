import { dummy } from './bar';
import * as fs from 'fs';
import { sync } from 'glob';
import { hasMagic } from 'glob';
export function dummyFs(fn: string) {
  fs.readFileSync(fn);
  return import('./bar');
}
export const dummy1 = dummy + 1;
export const readFile = fs.readFile;
export const globSync = sync;
export const hasMagic1 = hasMagic;
export { dummy2 } from './bar';
export * from './bar';
export { dummyBar2 } from './bar2';
