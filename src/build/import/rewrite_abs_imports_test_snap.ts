// Jest Snapshot v1, https://goo.gl/fbAQLP

exports[`rewrite_abs_imports should not rewrite relative paths 1`] = `
"import { DEPENDENCY } from \\"./dependency\\";
import * as dep from \\"./dependency\\";
export const DEP_PLUS_1 = DEPENDENCY + 1;
export const DEP_PLUS_100 = dep.DEPENDENCY + 100;"
`;

exports[
  `rewrite_abs_imports should rewrite absolute exports 1`
] = `"export { DEPENDENCY } from \\"./dependency\\";"`;

exports[
  `rewrite_abs_imports should rewrite absolute import expressions 1`
] = `"export const DYNAMIC_DEP = import(\\"./dependency\\").then(m => m.DEPENDENCY);"`;

exports[
  `rewrite_abs_imports should rewrite absolute import type paths in .d.ts 1`
] = `"export declare const IMPORT_TYPE: import(\\"./dependency\\");"`;

exports[`rewrite_abs_imports should rewrite absolute imports 1`] = `
"import { DEPENDENCY } from \\"./dependency\\";
import * as dep from \\"./dependency\\";
export const DEP_PLUS_1 = DEPENDENCY + 1;
export const DEP_PLUS_100 = dep.DEPENDENCY + 100;"
`;

exports[`rewrite_abs_imports should rewrite parent absolute paths 1`] = `
"import { DEPENDENCY } from \\"../dependency\\";
export const PARENT_DEP_1 = DEPENDENCY + 1;"
`;
