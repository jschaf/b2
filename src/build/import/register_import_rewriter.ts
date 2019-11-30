import { ImportRewriter } from './import_rewriter';
import * as fs from 'fs';
import * as jsonc from 'jsonc-parser';
import Module from 'module';
import * as path from 'path';

const findTsConfigRootDir = (startDir: string): string | null => {
  if (startDir === '') {
    throw new Error('Expected directory path but got empty string.');
  }
  let prevDir = '';
  let curDir = startDir;
  while (prevDir !== curDir) {
    const file = path.join(curDir, 'tsconfig.json');
    if (fs.existsSync(file)) {
      const conf = jsonc.parse(fs.readFileSync(file).toString('utf8'));
      const rootDir = conf['compilerOptions']['rootDir'];
      if (typeof rootDir !== 'string') {
        throw new Error(`Unable to parse rootDir from ${file}.`);
      }
      return path.resolve(curDir, rootDir);
    }

    prevDir = curDir;
    curDir = path.resolve(curDir, '..');
  }
  return null;
};

// Patch node's module loading
export const monkeyPatch = (): void => {
  const rootDir = findTsConfigRootDir(__dirname);
  if (rootDir === null) {
    throw new Error(
      `Unable to find package.json in ${__dirname} or any ` +
        `parent directory.`
    );
  }
  const importRewriter = ImportRewriter.forRootDir(rootDir);
  //@ts-ignore
  const origResolveFilename = Module._resolveFilename;
  //@ts-ignore
  Module._resolveFilename = function(
    request: string,
    parent: Module,
    _isMain: boolean,
    _options: unknown
  ): string {
    if (parent === null) {
      return request;
    }
    const rewrittenPath = importRewriter.rewrite(request, parent.filename);
    const rewrittenArgs = [rewrittenPath, ...Array.from(arguments).slice(1)];
    return origResolveFilename.apply(this, rewrittenArgs);
  };
};

monkeyPatch();
