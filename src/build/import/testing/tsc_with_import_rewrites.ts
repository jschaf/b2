import { checkDefined } from '//asserts';
import * as glob from 'glob';
import * as ts from 'typescript';
import * as rewriteAbsImport from '//build/import/rewrite_abs_imports';

const CJS_CONFIG: ts.CompilerOptions = {
  experimentalDecorators: true,
  jsx: ts.JsxEmit.React,
  module: ts.ModuleKind.ESNext,
  moduleResolution: ts.ModuleResolutionKind.NodeJs,
  noEmitOnError: false,
  noUnusedLocals: true,
  noUnusedParameters: true,
  stripInternal: true,
  declaration: true,
  baseUrl: __dirname,
  target: ts.ScriptTarget.ES2015,
};

export const compile = (
  input: string,
  opts: rewriteAbsImport.Opts,
  options: ts.CompilerOptions = CJS_CONFIG
): {} => {
  const files = glob.sync(input);
  const compilerHost = ts.createCompilerHost(options);
  const program = ts.createProgram(files, options, compilerHost);

  const msgs = {};

  const targetSourceFile = undefined;
  const writeFile = undefined;
  const cancellationToken = undefined;
  const emitOnlyDtsFiles = undefined;
  const customTransformers: ts.CustomTransformers = {
    after: [rewriteAbsImport.transformSourceFile(opts)],
    afterDeclarations: [rewriteAbsImport.transformBundleOrSourceFile(opts)],
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
