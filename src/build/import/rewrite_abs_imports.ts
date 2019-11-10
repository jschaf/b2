/**
 * AST Transformer to rewrite any ImportDeclaration paths.
 * This is typically used to rewrite relative imports into absolute imports
 * and mitigate import path differences.
 */
import {checkArg} from '//asserts';
import * as path from 'path';
import * as ts from 'typescript';
import {SyntaxKind} from 'typescript';

const ABS_PATH_PREFIX = '//';

/**
 * Rewrite relative import to absolute import or trigger
 * rewrite callback
 */
const rewritePath = (
  importPath: string,
  rootDir: string,
  sf: ts.SourceFile
): string => {
  if (!importPath.startsWith(ABS_PATH_PREFIX)) {
    return importPath;
  }
  // importPath: //back/db/orm
  // root: /home/code
  // sf.filename: /home/code/test/check/promises.ts

  // //back/db/orm => back/db/orm
  const relToRoot = importPath.slice(ABS_PATH_PREFIX.length);
  // back/db/orm => /home/code/back/db/orm
  const absImport = path.join(rootDir, relToRoot);
  // ../../back/db
  const relPath = path.relative(
      path.dirname(sf.fileName), path.dirname(absImport));
  const joined = path.join(relPath, path.basename(importPath));
  if (joined.startsWith('../') || joined.startsWith('./')) {
    return joined;
  } else {
    return `./${joined}`;
  }
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

const createAbsImportVisitor = (
  ctx: ts.TransformationContext,
  sf: ts.SourceFile,
  rootDir: string,
): ts.Visitor => {
  const visitor = (node: ts.Node): ts.Node => {
    // import $expr$ from $moduleSpecifier$;
    // export $expr$ from $moduleSpecifier$;
    if (ts.isImportDeclaration(node) || ts.isExportDeclaration(node)) {
      if (!node.moduleSpecifier) {
        return node;
      }
      const origPath = removeQuotes(node.moduleSpecifier.getText(sf));
      const rewrittenPath = rewritePath(origPath, rootDir, sf);
      const newNode = ts.getMutableClone(node);
      newNode.moduleSpecifier = ts.createLiteral(rewrittenPath);
      return newNode;
    }

    // const foo = import($arguments$);
    if (isDynamicImport(node)) {
      const origPath = removeQuotes(node.arguments[0].getText(sf));
      const rewrittenPath = rewritePath(origPath, rootDir, sf);
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
      const rewrittenPath = rewritePath(origPath, rootDir, sf);
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

export const transformBundleOrSourceFile = (
  projectBaseDir: string
): ts.TransformerFactory<ts.Bundle | ts.SourceFile> => {
  return (
    ctx: ts.TransformationContext
  ): ts.Transformer<ts.SourceFile | ts.Bundle> => {
    return (sf: ts.SourceFile | ts.Bundle) => {
      if (sf.kind !== SyntaxKind.SourceFile) {
        throw new Error('Only SourceFile transform supported');
      }
      return ts.visitNode(sf, createAbsImportVisitor(ctx, sf, projectBaseDir));
    };
  };
};

export const transformSourceFile = (
    projectBaseDir: string
): ts.TransformerFactory<ts.SourceFile> => {
  return (ctx: ts.TransformationContext): ts.Transformer<ts.SourceFile> => {
    return (sf: ts.SourceFile) =>
      ts.visitNode(sf, createAbsImportVisitor(ctx, sf, projectBaseDir));
  };
};
