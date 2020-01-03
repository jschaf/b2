/** Compiles a post into HTML on top of a mempost. */
import { HastCompiler } from '//post/hast/compiler';
import { MdastCompiler } from '//post/mdast/compiler';
import { Mempost } from '//post/mempost';
import { PostAST } from '//post/post_ast';

/** Compiles a post AST into a mempost ready to saved to a file system. */
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
    const hastNode = this.mdastCompiler.compile(postAST);
    const html = this.hastCompiler.compile(hastNode);
    const dest = Mempost.create();
    dest.addUtf8Entry('index.html', html);
    return dest;
  }
}
