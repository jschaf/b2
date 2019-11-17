import * as path from 'path';

const ABS_PATH_PREFIX = '//';

export class ImportRewriter {
  private constructor(private readonly rootDir: string) {
  }

  static forRootDir(dir: string): ImportRewriter {
    return new ImportRewriter(dir);
  }

  /**
   * Rewrite relative import to absolute import or trigger
   * rewrite callback
   */
  rewrite(importPath: string, parent: string): string {
    if (!importPath.startsWith(ABS_PATH_PREFIX)) {
      return importPath;
    }
    const relToRoot = importPath.slice(ABS_PATH_PREFIX.length);
    const absImport = path.join(this.rootDir, relToRoot);
    const relPath = path.relative(
        path.dirname(parent),
        path.dirname(absImport)
    );
    const joined = path.join(relPath, path.basename(importPath));
    if (joined.startsWith('../') || joined.startsWith('./')) {
      return joined;
    } else {
      // When both the import path and parent are in the same dir
      // the relPath is something like `foo` so make it an explicitly
      // relative path.
      return `./${joined}`;
    }
  }
}

