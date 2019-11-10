import {checkDefined} from '//asserts';
import * as files from '//files';
import * as tscImportRewrite from '//build/import/testing/tsc_with_import_rewrites';
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
    // Needed to use ts.sys functions without creating an arrow function
    // for each one.
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

    const outDir = checkDefined(tsConfig.options.outDir);
    files.deleteDirectory(path.resolve(baseDir, outDir));
    tscImportRewrite.compile(rootNames, baseDir, tsConfig.options);
    /* eslint-enable */
  });

  it('should not rewrite relative paths', async () => {
    const content = await readRelativeFile('testdata/esnext_proj/out/rel_import.js');
    expect(content).toMatchSnapshot();
  });

  it('should rewrite absolute imports', async () => {
    const content = await readRelativeFile('testdata/esnext_proj/out/abs_import.js');
    expect(content).toMatchSnapshot();
  });

  it('should rewrite absolute exports', async () => {
    const content = await readRelativeFile('testdata/esnext_proj/out/abs_export.js');
    expect(content).toMatchSnapshot();
  });

  it('should rewrite parent absolute paths', async () => {
    const content = await readRelativeFile('testdata/esnext_proj/out/child/parent_import.js');
    expect(content).toMatchSnapshot();
  });

  it('should rewrite absolute import type paths in .d.ts', async () => {
    const content = await readRelativeFile('testdata/esnext_proj/out/abs_import_type.d.ts');
    expect(content).toMatchSnapshot();
  });

  it('should rewrite absolute import expressions', async () => {
    const content = await readRelativeFile('testdata/esnext_proj/out/abs_import_expr.js');
    expect(content).toMatchSnapshot();
  });
});
