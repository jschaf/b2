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

/** Finds the nearest parent directory that contains a .git folder. */
export const findGitDirectory = (dir: string): string => {
  checkArg(dir.length > 0, 'Directory cannot be empty.');
  const parents = dir.split(path.sep);
  while (parents.length > 0) {
    const dir = parents.join(path.sep);
    const gitPath = path.join(dir, '.git');
    if (fs.existsSync(gitPath)) {
      return path.normalize(gitPath);
    }
    parents.pop();
  }
  throw new Error(`Unable to find .git dir in any parent starting from ${dir}`);
};
