import { checkArg } from '//asserts';
import * as fs from 'fs';
import * as path from 'path';

/** Deletes a directory recursively. */
export const deleteDirectory = (dir: string): void => {
  if (!fs.existsSync(dir)) {
    return;
  }
  checkArg(path.resolve(dir) !== '/', 'Not deleting root');

  for (const child of fs.readdirSync(dir)) {
    const curPath = path.join(dir, child);
    if (fs.lstatSync(curPath).isDirectory()) {
      deleteDirectory(curPath);
    } else {
      fs.unlinkSync(curPath);
    }
  }
  fs.rmdirSync(dir);
};
