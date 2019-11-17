import { checkDefined } from '//asserts';
import * as rewriteAbsImport from '//build/import/rewrite_abs_imports';
import { PluggableTsc, TsConfigParser } from '//build/pluggable_tsc';
import * as files from '//files';
import * as fs from 'fs';
import * as path from 'path';

const resolveRelativeFile = (filePath: string): string => {
  return path.resolve(__dirname, filePath);
};

const readRelativeFile = async (filePath: string): Promise<string> => {
  const buffer = await fs.promises.readFile(resolveRelativeFile(filePath));
  return buffer.toString('utf8').trim();
};

describe('rewrite_abs_imports', () => {
  beforeAll(() => {
    const baseDir = path.resolve(__dirname, 'testdata/esnext_proj');
    const rootNames = [
      resolveRelativeFile('testdata/esnext_proj/rel_import.ts'),
      resolveRelativeFile('testdata/esnext_proj/abs_import.ts'),
      resolveRelativeFile('testdata/esnext_proj/abs_export.ts'),
      resolveRelativeFile('testdata/esnext_proj/abs_import_type.ts'),
      resolveRelativeFile('testdata/esnext_proj/abs_import_expr.ts'),
      resolveRelativeFile('testdata/esnext_proj/child_import.ts'),
      resolveRelativeFile('testdata/esnext_proj/child/parent_import.ts'),
    ];

    const tsConfig = TsConfigParser.forDirectory(baseDir).parse();
    files.deleteDirectory(
      path.resolve(baseDir, checkDefined(tsConfig.options.outDir))
    );
    const tsc = PluggableTsc.forOptions(tsConfig);
    tsc.addAfterTransformer(rewriteAbsImport.newAfterTransformer(baseDir));
    tsc.addAfterDeclarationsTransformer(
      rewriteAbsImport.newAfterDeclarationsTransformer(baseDir)
    );
    tsc.compile(rootNames);
  });

  it('should not rewrite relative paths', async () => {
    const content = await readRelativeFile(
      'testdata/esnext_proj/out/rel_import.js'
    );
    expect(content).toMatchSnapshot();
  });

  it('should rewrite absolute imports', async () => {
    const content = await readRelativeFile(
      'testdata/esnext_proj/out/abs_import.js'
    );
    expect(content).toMatchSnapshot();
  });

  it('should rewrite absolute exports', async () => {
    const content = await readRelativeFile(
      'testdata/esnext_proj/out/abs_export.js'
    );
    expect(content).toMatchSnapshot();
  });

  it('should rewrite parent absolute paths', async () => {
    const content = await readRelativeFile(
      'testdata/esnext_proj/out/child/parent_import.js'
    );
    expect(content).toMatchSnapshot();
  });

  it('should rewrite absolute import type paths in .d.ts', async () => {
    const content = await readRelativeFile(
      'testdata/esnext_proj/out/abs_import_type.d.ts'
    );
    expect(content).toMatchSnapshot();
  });

  it('should rewrite absolute import expressions', async () => {
    const content = await readRelativeFile(
      'testdata/esnext_proj/out/abs_import_expr.js'
    );
    expect(content).toMatchSnapshot();
  });
});
