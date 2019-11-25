import * as path from 'path';
import {checkDefined} from '//asserts';
import * as rewriteAbsImport from '//build/import/rewrite_abs_imports';
import * as ts from 'typescript';

const tsNode = require('ts-node').register;

/** Parses the tsconfig file from a directory. */
export class TsConfigParser {

  private readonly parseHost: ts.ParseConfigHost = {
    fileExists: ts.sys.fileExists,
    readFile: ts.sys.readFile,
    readDirectory: ts.sys.readDirectory,
    useCaseSensitiveFileNames: true,
  };

  private constructor(private readonly dir: string) {
  }

  static forCurrentProject(): TsConfigParser {
    return TsConfigParser.forDirectory(process.cwd());
  }

  static forDirectory(dir: string): TsConfigParser {
    return new TsConfigParser(dir);
  }

  parseRaw(): any {
    const config = checkDefined(ts.findConfigFile(this.dir, ts.sys.fileExists));
    const sourceFile = ts.readJsonConfigFile(config, ts.sys.readFile);
    const errors: ts.Diagnostic[] = [];
    const tsConf = ts.convertToObject(sourceFile, errors);
    if (errors.length > 0) {
      throw new Error(`Had errors parsing tsconfig.json: ${errors.join('\n')}`);
    }
    return tsConf;
  }

  parse(): ts.ParsedCommandLine {
    /* eslint-disable @typescript-eslint/unbound-method */
    // Needed to use ts.sys functions without creating an arrow function
    // for each one.
    const config = checkDefined(ts.findConfigFile(this.dir, ts.sys.fileExists));
    const sourceFile = ts.readJsonConfigFile(config, ts.sys.readFile);
    /* eslint-enable */
    return ts.parseJsonSourceFileConfigFileContent(
        sourceFile,
        this.parseHost,
        this.dir
    );
  }
}

type TscTransformer<T extends ts.Node> =
    | ts.TransformerFactory<T>
    | ts.CustomTransformerFactory;

export class PluggableTsc {
  private readonly beforeTransfomers: TscTransformer<ts.SourceFile>[] = [];
  private readonly afterTransfomers: TscTransformer<ts.SourceFile>[] = [];
  private readonly afterDeclarationsTransfomers: TscTransformer<ts.SourceFile | ts.Bundle>[] = [];

  private constructor(private readonly opts: ts.ParsedCommandLine) {
  }

  static forOptions(opts: ts.ParsedCommandLine): PluggableTsc {
    return new PluggableTsc(opts);
  }

  addBeforeTransformer(t: TscTransformer<ts.SourceFile>): void {
    this.beforeTransfomers.push(t);
  }

  addAfterTransformer(t: TscTransformer<ts.SourceFile>): void {
    this.afterTransfomers.push(t);
  }

  addAfterDeclarationsTransformer(
      t: TscTransformer<ts.SourceFile | ts.Bundle>
  ): void {
    this.afterDeclarationsTransfomers.push(t);
  }

  compile(rootNames: string[]): void {
    const compilerHost = ts.createCompilerHost(this.opts.options);
    const program = ts.createProgram(
        rootNames,
        this.opts.options,
        compilerHost
    );

    const targetSourceFile = undefined;
    const writeFile = undefined;
    const cancellationToken = undefined;
    const emitOnlyDtsFiles = undefined;
    const customTransformers: ts.CustomTransformers = {
      before: this.beforeTransfomers,
      after: this.afterTransfomers,
      afterDeclarations: this.afterDeclarationsTransfomers,
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
      const {line, character} = file.getLineAndCharacterOfPosition(
          checkDefined(diagnostic.start)
      );
      const message = ts.flattenDiagnosticMessageText(
          diagnostic.messageText,
          '\n'
      );
      console.log(
          `${file.fileName} (${line + 1},${character + 1}): ${message}`
      );
    }
  }
}

if (require.main === module) {
  const opts = TsConfigParser.forCurrentProject().parse();
  const dir = checkDefined(opts.options.baseUrl);
  const roots = process.argv.slice(2);
  const rewriteAbsImportAfter = rewriteAbsImport.newAfterTransformer(dir);
  const rewriteAbsImportAfterDecl = rewriteAbsImport.newAfterDeclarationsTransformer(dir);

  if (roots.length > 0) {
    const rawTsConf = TsConfigParser.forCurrentProject().parseRaw();
    tsNode({
      files: roots, compilerOptions: rawTsConf.compilerOptions, transformers: {
        after: [rewriteAbsImportAfter],
        afterDeclarations: [rewriteAbsImportAfterDecl]
      }
    });
    for (const root of roots) {
      const absPath = path.resolve(process.cwd(), root);
      const pp = path.parse(absPath);
      const stripped = path.join(pp.dir, pp.name);
      const relPath = path.relative(process.cwd(), stripped);
      if (relPath.startsWith('./') || relPath.startsWith('../')) {
        require(relPath);
      } else {
        require('./' + relPath);
      }
    }

  } else {
    const tsc = PluggableTsc.forOptions(opts);
    tsc.addAfterTransformer(rewriteAbsImportAfter);
    tsc.addAfterDeclarationsTransformer(rewriteAbsImportAfterDecl);
    tsc.compile(roots);
  }
}
