import {checkDefined} from '//asserts';
import * as tscImportRewrite from '//build/import/testing/tsc_with_import_rewrites';
import {dedent} from '//strings';
import * as fs from 'fs';
import * as path from 'path';
import * as ts from 'typescript';

const resolveRelativeFile = (filePath: string): string => {
  return path.resolve(__dirname, filePath);
};

const readRelativeFile = async (filePath: string): Promise<string> => {
  const buffer = await fs.promises.readFile(resolveRelativeFile(filePath));
  return buffer.toString('utf8').trim();
};

describe('rewrite_abs_imports', () => {

  beforeAll(() => {
    /* eslint-disable @typescript-eslint/unbound-method */
    const parseConfigHost: ts.ParseConfigHost = {
      fileExists: ts.sys.fileExists,
      readFile: ts.sys.readFile,
      readDirectory: ts.sys.readDirectory,
      useCaseSensitiveFileNames: true
    };
    const baseDir = path.resolve(__dirname, 'testdata/esnext_proj');
    const config = checkDefined(ts.findConfigFile(baseDir, ts.sys.fileExists));
    const sourceFile = ts.readJsonConfigFile(config, ts.sys.readFile);
    const tsConfig = ts.parseJsonSourceFileConfigFileContent(sourceFile, parseConfigHost, baseDir);
    const rootNames = [
      resolveRelativeFile('testdata/esnext_proj/rel_import.ts'),
      resolveRelativeFile('testdata/esnext_proj/abs_import.ts'),
      resolveRelativeFile('testdata/esnext_proj/abs_export.ts'),
      resolveRelativeFile('testdata/esnext_proj/abs_import_type.ts'),
      resolveRelativeFile('testdata/esnext_proj/abs_import_expr.ts'),
      resolveRelativeFile('testdata/esnext_proj/child_import.ts'),
      resolveRelativeFile('testdata/esnext_proj/child/parent_import.ts'),
    ];
    tscImportRewrite.compile(rootNames, baseDir, tsConfig.options);
    /* eslint-enable */
  });

  it('should not rewrite relative paths', async () => {
    const content = await readRelativeFile('testdata/esnext_proj/out/rel_import.js');

    expect(content).toEqual(dedent`
      import { DEPENDENCY } from "./dependency";
      import * as dep from "./dependency";
      export const DEP_PLUS_1 = DEPENDENCY + 1;
      export const DEP_PLUS_100 = dep.DEPENDENCY + 100;
    `);
  });

  it('should rewrite absolute imports', async () => {
    const content = await readRelativeFile('testdata/esnext_proj/out/abs_import.js');

    expect(content).toEqual(dedent`
      import { DEPENDENCY } from "./dependency";
      import * as dep from "./dependency";
      export const DEP_PLUS_1 = DEPENDENCY + 1;
      export const DEP_PLUS_100 = dep.DEPENDENCY + 100;
    `);
  });

  it('should rewrite absolute exports', async () => {
    const content = await readRelativeFile('testdata/esnext_proj/out/abs_export.js');

    expect(content).toEqual(dedent`
      export { DEPENDENCY } from "./dependency";
    `);
  });

  it('should rewrite parent absolute paths', async () => {
    const content = await readRelativeFile('testdata/esnext_proj/out/child/parent_import.js');

    expect(content).toEqual(dedent`
      import { DEPENDENCY } from "../dependency";
      export const PARENT_DEP_1 = DEPENDENCY + 1;
    `);
  });

  it('should rewrite absolute import type paths in .d.ts', async () => {
    const content = await readRelativeFile('testdata/esnext_proj/out/abs_import_type.d.ts');

    expect(content).toEqual(dedent`
      export declare const IMPORT_TYPE: import("./dependency");
    `);
  });

  it('should rewrite absolute import expressions', async () => {
    const content = await readRelativeFile('testdata/esnext_proj/out/abs_import_expr.js');

    expect(content).toEqual(dedent`
      export const DYNAMIC_DEP = import("./dependency").then(m => m.DEPENDENCY);
    `);
  });

//  it('should produce expected js output', async () => {
//    const content = await readRelativeFile('testdata/foo.js');
//    expect(content).toEqual(dedent`
//        import { dummy } from "dummy-project/import/testdata/bar";
//        import * as fs from "rewritten/fs";
//        import { sync } from "external/glob";
//        import { hasMagic } from "external/glob";
//        export function dummyFs(fn) {
//            fs.readFileSync(fn);
//            return import("dummy-project/import/testdata/bar");
//        }
//        export const dummy1 = dummy + 1;
//        export const readFile = fs.readFile;
//        export const globSync = sync;
//        export const hasMagic1 = hasMagic;
//        export { dummy2 } from "dummy-project/import/testdata/bar";
//        export * from "dummy-project/import/testdata/bar";
//        export { dummyBar2 } from "relative";
//      `);
//  });
//  it('should produce expected d.ts output', async () => {
//    const content = await readRelativeFile('testdata/foo.d.ts');
//    expect(content).toContain(
//      'export { dummy2 } from "dummy-project/import/testdata/bar";'
//    );
//    expect(content).toContain(
//      'export * from "dummy-project/import/testdata/bar";'
//    );
//  });
//
//  it('should prioritize alias over relative resolution', async () => {
//    const content = await readRelativeFile('testdata/foo.d.ts');
//    expect(content).toContain('export { dummyBar2 } from "relative";');
//  });
});
