import * as path from 'path';
import {ImportRewriter} from '//build/import/import_rewriter';

const testRootDir = (subPath: string = '.'): string => {
  return path.join('/root/src/', subPath);
};

const rewrite = (path: string, parentFile: string): string => {
  const ir = ImportRewriter.forRootDir(testRootDir());
  return ir.rewrite(path, testRootDir(parentFile));
};

describe('ImportRewriter', () => {
  it('should rewrite paths in same dir', () => {
    expect(rewrite('//foo', 'parent')).toEqual('./foo');
  });

  it('should rewrite paths to a parent', () => {
    expect(rewrite('//foo', 'parent/child')).toEqual('../foo');
  });

  it('should rewrite paths to a child', () => {
    expect(rewrite('//foo/bar', 'parent')).toEqual('./foo/bar');
  });
});
