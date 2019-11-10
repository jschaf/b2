import { checkDefined } from '//asserts';
import * as ts from 'typescript';
import * as rewriteAbsImport from '//build/import/rewrite_abs_imports';

export const compile = (
  rootNames: string[],
  options: ts.CompilerOptions
): {} => {
  const compilerHost = ts.createCompilerHost(options);
  const program = ts.createProgram(rootNames, options, compilerHost);
  const baseDir = checkDefined(options.baseUrl);

  const msgs = {};

  const targetSourceFile = undefined;
  const writeFile = undefined;
  const cancellationToken = undefined;
  const emitOnlyDtsFiles = undefined;
  const customTransformers: ts.CustomTransformers = {
    after: [rewriteAbsImport.transformSourceFile(baseDir)],
    afterDeclarations: [rewriteAbsImport.transformBundleOrSourceFile(baseDir)],
  };
  const emitResult = program.emit(
    targetSourceFile,
    writeFile,
    cancellationToken,
    emitOnlyDtsFiles,
    customTransformers
  );

  const allDiagnostics = ts
    .getPreEmitDiagnostics(program)
    .concat(emitResult.diagnostics);

  for (const diagnostic of allDiagnostics) {
    const file = checkDefined(diagnostic.file);
    const { line, character } = file.getLineAndCharacterOfPosition(
      checkDefined(diagnostic.start)
    );
    const message = ts.flattenDiagnosticMessageText(
      diagnostic.messageText,
      '\n'
    );
    console.log(`${file.fileName} (${line + 1},${character + 1}): ${message}`);
  }

  return msgs;
};
