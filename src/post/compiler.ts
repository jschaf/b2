/** Compiles a post into HTML on top of a mempost. */
import { checkState } from '//asserts';
import { HastCompiler } from '//post/hast/compiler';
import { MdastCompiler } from '//post/mdast/compiler';
import * as md from '//post/mdast/nodes';
import { Mempost } from '//post/mempost';
import { PostAST } from '//post/ast';

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

  compileToMempost(postAST: PostAST): Mempost {
    checkState(md.isRoot(postAST.mdastNode), 'Post AST node must be root node');
    const hastNode = this.mdastCompiler.compile(postAST);
    checkState(hastNode.length === 1, 'Expected exactly 1 hast node');
    const html = this.hastCompiler.compile(hastNode[0]);
    const dest = Mempost.create();
    dest.addUtf8Entry('index.html', html);
    return dest;
  }
}
