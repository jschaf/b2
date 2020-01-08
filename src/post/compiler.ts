/** Compiles a post into HTML on top of a mempost. */
import { checkState } from '//asserts';
import { HastCompiler } from '//post/hast/compiler';
import { MdastCompiler } from '//post/mdast/compiler';
import * as md from '//post/mdast/nodes';
import { Mempost } from '//post/mempost';
import { PostAST } from '//post/ast';
import * as fs from 'fs';
import * as path from 'path';

/** Compiles a post AST into a mempost ready to be saved to a file system. */
export class PostCompiler {
  private constructor(
    private readonly mdastCompiler: MdastCompiler,
    private readonly hastCompiler: HastCompiler
  ) {}

  static create(): PostCompiler {
    return new PostCompiler(
      MdastCompiler.createDefault(),
      HastCompiler.create()
    );
  }

  compile(postAST: PostAST): CompiledPost {
    checkState(md.isRoot(postAST.mdastNode), 'Post AST node must be root node');
    const hastNode = this.mdastCompiler.compile(postAST.mdastNode, postAST);
    checkState(hastNode.length === 1, 'Expected exactly 1 hast node');
    const html = this.hastCompiler.compile(hastNode[0], postAST);
    const dest = Mempost.create();
    dest.addUtf8Entry('index.html', html);
    return CompiledPost.create(postAST, dest);
  }
}

export class CompiledPost {
  private constructor(readonly ast: PostAST, readonly mempost: Mempost) {}

  static create(ast: PostAST, mempost: Mempost): CompiledPost {
    return new CompiledPost(ast, mempost);
  }

  async write(destDir: string): Promise<void> {
    const p = path.join(destDir, this.ast.metadata.slug, 'index.html');
    await fs.promises.mkdir(path.dirname(p), { recursive: true });
    await fs.promises.writeFile(p, this.mempost.getEntry('index.html'));
  }
}
