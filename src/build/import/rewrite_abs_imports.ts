/**
 * AST Transformer to rewrite any ImportDeclaration paths.
 * This is typically used to rewrite relative imports into absolute imports
 * and mitigate import path differences.
 */
import { checkArg, checkState } from '//asserts';
import { dirname, resolve } from 'path';
import * as ts from 'typescript';
import { SyntaxKind } from 'typescript';

export interface Opts {
  projectBaseDir?: string;
  project: string;

  rewrite(importPath: string, sourceFilePath: string): string | undefined;

  alias: Record<string, string>;
}

/**
 * Rewrite relative import to absolute import or trigger
 * rewrite callback
 */
const rewritePath = (
  importPath: string,
  sf: ts.SourceFile,
  opts: Opts,
  regexps: Record<string, RegExp>
): string => {
  const aliases = Object.keys(regexps);
  for (const alias of aliases) {
    const regex = regexps[alias];
    if (regexps[alias].test(importPath)) {
      return importPath.replace(regex, opts.alias[alias]);
    }
  }

  checkState(typeof opts.rewrite === 'function');
  const newImportPath = opts.rewrite(importPath, sf.fileName);
  if (newImportPath) {
    return newImportPath;
  }

  if (opts.project && opts.projectBaseDir && importPath.startsWith('.')) {
    const path = resolve(dirname(sf.fileName), importPath).split(
      opts.projectBaseDir
    )[1];
    return `${opts.project}${path}`;
  }

  return importPath;
};

const isDynamicImport = (node: ts.Node): node is ts.CallExpression => {
  return (
    ts.isCallExpression(node) &&
    node.expression.kind === ts.SyntaxKind.ImportKeyword
  );
};

const removeQuotes = (text: string): string => {
  checkArg(text.length >= 2);
  checkArg(text.startsWith("'") || text.startsWith("'"));
  checkArg(text.endsWith("'") || text.endsWith("'"));
  return text.substr(1, text.length - 2);
};

const importExportVisitor = (
  ctx: ts.TransformationContext,
  sf: ts.SourceFile,
  opts: Opts,
  regexps: Record<string, RegExp>
): ts.Visitor => {
  const visitor = (node: ts.Node): ts.Node => {
    // import $expr$ from $moduleSpecifier$;
    // export $expr$ from $moduleSpecifier$;
    if (ts.isImportDeclaration(node) || ts.isExportDeclaration(node)) {
      if (!node.moduleSpecifier) {
        return node;
      }
      const origPath = removeQuotes(node.moduleSpecifier.getText(sf));
      const rewrittenPath = rewritePath(origPath, sf, opts, regexps);
      const newNode = ts.getMutableClone(node);
      newNode.moduleSpecifier = ts.createLiteral(rewrittenPath);
      return newNode;
    }

    // const foo = import($arguments$);
    if (isDynamicImport(node)) {
      const origPath = removeQuotes(node.arguments[0].getText(sf));
      const rewrittenPath = rewritePath(origPath, sf, opts, regexps);
      const newNode = ts.getMutableClone(node);
      newNode.arguments = ts.createNodeArray([
        ts.createStringLiteral(rewrittenPath),
      ]);
      return newNode;
    }

    // declare const foo: import($stringLiteral$);
    if (
        ts.isImportTypeNode(node) &&
        ts.isLiteralTypeNode(node.argument) &&
        ts.isStringLiteral(node.argument.literal)
    ) {
      // `.text` instead of `getText` because this node doesn't map to sf. It's
      // a generated d.ts file.
      const origPath = node.argument.literal.text;
      const rewrittenPath = rewritePath(origPath, sf, opts, regexps);
      const newNode = ts.getMutableClone(node);
      newNode.argument = ts.createLiteralTypeNode(
          ts.createStringLiteral(rewrittenPath)
      );
      return newNode;
    }

    // Everything else.
    return ts.visitEachChild(node, visitor, ctx);
  };
  return visitor;
};

const buildTransformRegexps = (
  alias: Record<string, string>
): Record<string, RegExp> => {
  return Object.keys(alias).reduce(
    (all, regexString) => {
      all[regexString] = new RegExp(regexString, 'gi');
      return all;
    },
    {} as Record<string, RegExp>
  );
};

export const transformBundleOrSourceFile = (
  opts: Opts
): ts.TransformerFactory<ts.Bundle | ts.SourceFile> => {
  return (
    ctx: ts.TransformationContext
  ): ts.Transformer<ts.SourceFile | ts.Bundle> => {
    return (sf: ts.SourceFile | ts.Bundle) => {
      if (sf.kind !== SyntaxKind.SourceFile) {
        throw new Error('Only SourceFile transform supported');
      }
      const regexps = buildTransformRegexps(opts.alias);
      return ts.visitNode(sf, importExportVisitor(ctx, sf, opts, regexps));
    };
  };
};

export const transformSourceFile = (
  opts: Opts
): ts.TransformerFactory<ts.SourceFile> => {
  const { alias = {} } = opts;
  const regexps = buildTransformRegexps(alias);
  return (ctx: ts.TransformationContext): ts.Transformer<ts.SourceFile> => {
    return (sf: ts.SourceFile) =>
      ts.visitNode(sf, importExportVisitor(ctx, sf, opts, regexps));
  };
};
