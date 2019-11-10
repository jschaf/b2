import * as rewriteAbsImports from '//build/import/rewrite_abs_imports';
import * as tscImportRewrite from '//build/import/testing/tsc_with_import_rewrites';
import { dedent } from '//strings';
import * as fs from 'fs';
import * as path from 'path';
import { resolve } from 'path';

const opts: rewriteAbsImports.Opts = {
  projectBaseDir: resolve(__dirname, '..'),
  project: 'dummy-project',
  rewrite(importPath) {
    if (importPath.startsWith('fs')) {
      return 'rewritten/fs';
    } else if (importPath.startsWith('//')) {
      return '.*/foo';
    } else {
      return undefined;
    }
  },
  alias: {
    './bar2': 'relative',
    '^(glob)$': 'external/$1',
  },
};

const readRelativeFile = async (testDataPath: string): Promise<string> => {
  const buffer = await fs.promises.readFile(
    path.resolve(__dirname, testDataPath)
  );
  return buffer.toString('utf8');
};

describe('rewrite_abs_imports', () => {
  beforeAll(() => {
    tscImportRewrite.compile(path.resolve(__dirname, 'testdata/foo.ts'), opts);
  });

  it('should produce expected js output', async () => {
    const content = (await readRelativeFile('testdata/foo.js')).trim();
    expect(content).toEqual(dedent`
        import { dummy } from "dummy-project/import/testdata/bar";
        import * as fs from "rewritten/fs";
        import { sync } from "external/glob";
        import { hasMagic } from "external/glob";
        export function dummyFs(fn) {
            fs.readFileSync(fn);
            return import("dummy-project/import/testdata/bar");
        }
        export const dummy1 = dummy + 1;
        export const readFile = fs.readFile;
        export const globSync = sync;
        export const hasMagic1 = hasMagic;
        export { dummy2 } from "dummy-project/import/testdata/bar";
        export * from "dummy-project/import/testdata/bar";
        export { dummyBar2 } from "relative";
      `);
  });
  it('should produce expected d.ts output', async () => {
    const content = await readRelativeFile('testdata/foo.d.ts');
    expect(content).toContain(
      'export { dummy2 } from "dummy-project/import/testdata/bar";'
    );
    expect(content).toContain(
      'export * from "dummy-project/import/testdata/bar";'
    );
  });

  it('should prioritize alias over relative resolution', async () => {
    const content = await readRelativeFile('testdata/foo.d.ts');
    expect(content).toContain('export { dummyBar2 } from "relative";');
  });
});
