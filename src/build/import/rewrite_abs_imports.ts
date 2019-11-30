/**
 * AST Transformer to rewrite any ImportDeclaration paths.
 * This is typically used to rewrite relative imports into absolute imports
 * and mitigate import path differences.
 */
import { checkArg, checkDefined } from '//asserts';
import { ImportRewriter } from '//build/import/import_rewriter';
import * as ts from 'typescript';
import { SyntaxKind } from 'typescript';

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
  rootDir: string
): ts.Visitor => {
  const importRewriter = ImportRewriter.forRootDir(rootDir);
  const visitor = (node: ts.Node): ts.Node => {
    // import $expr$ from $moduleSpecifier$;
    // export $expr$ from $moduleSpecifier$;
    if (ts.isImportDeclaration(node) || ts.isExportDeclaration(node)) {
      if (!node.moduleSpecifier) {
        return node;
      }
      const origPath = removeQuotes(node.moduleSpecifier.getText(sf));
      const rewrittenPath = importRewriter.rewrite(origPath, sf.fileName);
      const newNode = ts.getMutableClone(node);
      newNode.moduleSpecifier = ts.createLiteral(rewrittenPath);
      return newNode;
    }

    // const foo = import($arguments$);
    if (isDynamicImport(node)) {
      const origPath = removeQuotes(node.arguments[0].getText(sf));
      const rewrittenPath = importRewriter.rewrite(origPath, sf.fileName);
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
      const rewrittenPath = importRewriter.rewrite(origPath, sf.fileName);
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

export const newAfterDeclarationsTransformer = (
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

export const newAfterTransformer = (
  projectBaseDir: string
): ts.TransformerFactory<ts.SourceFile> => {
  return (ctx: ts.TransformationContext): ts.Transformer<ts.SourceFile> => {
    return (sf: ts.SourceFile) =>
      ts.visitNode(sf, createAbsImportVisitor(ctx, sf, projectBaseDir));
  };
};

export const transform = (
  program: ts.Program,
  _pluginOptions: {}
): ts.TransformerFactory<ts.SourceFile> => {
  const projectBaseDir = checkDefined(program.getCompilerOptions().baseUrl);
  return newAfterTransformer(projectBaseDir);
};
