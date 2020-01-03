import { checkDefined, checkState } from '//asserts';
import { MdastCompiler } from '//post/mdast/compiler';
import * as h from '//post/hast/nodes';
import * as md from '//post/mdast/nodes';
import { PostAST } from '//post/post_ast';
import { isString } from '//strings';
import * as mdast from 'mdast';
import * as unist from 'unist';
import * as unistNodes from '//unist/nodes';

/** Compiler for a single mdast node. */
export interface MdastNodeCompiler {
  compileNode(node: unist.Node, ast: PostAST): unist.Node;
}

/**
 * Compiles an mdast blockquote to hast, like:
 *
 *     > foo bar
 *     > baz qux
 *
 * https://github.com/syntax-tree/mdast#blockquote
 */
export class BlockquoteCompiler implements MdastNodeCompiler {
  private constructor(private readonly compiler: MdastCompiler) {}

  static create(compiler: MdastCompiler): BlockquoteCompiler {
    return new BlockquoteCompiler(compiler);
  }

  compileNode(node: unist.Node, postAST: PostAST): unist.Node {
    md.checkType(node, 'blockquote', md.isBlockquote);
    const children = this.compiler.compileChildren(node, postAST);
    return h.elem('blockquote', children);
  }
}

/**
 * Compiles an mdast break to hast, like using two spaces before a newline.
 *
 * https://github.com/syntax-tree/mdast#break
 */
export class BreakCompiler implements MdastNodeCompiler {
  private constructor() {}

  static create(): BreakCompiler {
    return new BreakCompiler();
  }

  compileNode(node: unist.Node, _postAST: PostAST): unist.Node {
    md.checkType(node, 'break', md.isBreak);
    return h.elem('break');
  }
}

/**
 * Compiles an mdast code block to hast, like:
 *
 *     let f = () => 42;
 *
 * https://github.com/syntax-tree/mdast#code
 */
export class CodeCompiler implements MdastNodeCompiler {
  private constructor() {}

  static create(): CodeCompiler {
    return new CodeCompiler();
  }

  compileNode(node: unist.Node, _postAST: PostAST): unist.Node {
    md.checkType(node, 'code', md.isCode);
    const props: Record<string, unknown> = {};
    if (isString(node.lang) && node.lang !== '') {
      props.className = ['lang-' + node.lang.trim()];
    }
    return h.elem('pre', [h.elemProps('code', props, [h.text(node.value)])]);
  }
}

/**
 * Compiles deleted mdast text to hast, like:
 *
 *     Lorem ipsum ~~this is deleted~~.
 *
 * https://github.com/syntax-tree/mdast#delete
 */
export class DeleteCompiler implements MdastNodeCompiler {
  private constructor(private readonly compiler: MdastCompiler) {}

  static create(compiler: MdastCompiler): DeleteCompiler {
    return new DeleteCompiler(compiler);
  }

  compileNode(node: unist.Node, postAST: PostAST): unist.Node {
    md.checkType(node, 'delete', md.isDelete);
    const children = this.compiler.compileChildren(node, postAST);
    return h.elem('del', children);
  }
}

/**
 * Compiles emphasized mdast text to hast, like:
 *
 *     Foo bar *this is emphasized* and _so is this_.
 *
 * https://github.com/syntax-tree/mdast#emphasis
 */
export class EmphasisCompiler implements MdastNodeCompiler {
  private constructor(private readonly compiler: MdastCompiler) {}

  static create(compiler: MdastCompiler): EmphasisCompiler {
    return new EmphasisCompiler(compiler);
  }

  compileNode(node: unist.Node, postAST: PostAST): unist.Node {
    md.checkType(node, 'emphasis', md.isEmphasis);
    const children = this.compiler.compileChildren(node, postAST);
    return h.elem('em', children);
  }
}

/**
 * Compiles an mdast inline footnote definition to hast, like:
 *
 *      Foo bar [^inline footnote] qux.
 *
 * https://github.com/syntax-tree/mdast#footnote
 */
export class FootnoteCompiler implements MdastNodeCompiler {
  private constructor(private readonly compiler: MdastCompiler) {}

  static create(compiler: MdastCompiler): FootnoteCompiler {
    return new FootnoteCompiler(compiler);
  }

  compileNode(node: unist.Node, postAST: PostAST): unist.Node {
    md.checkType(node, 'footnote', md.isFootnote);
    const data = checkDefined(
      node.data,
      'Expected data attr to exist on footnote type.'
    );
    let key = PostAST.INLINE_FOOTNOTE_DATA_KEY;
    const fnId = checkDefined(
      data[key],
      `Expected a footnote ID to exist on data attr with key: ${key}`
    );
    checkState(
      isString(fnId),
      `Expected data.${key} to be a string but was ${fnId}`
    );
    const fnRef = md.footnoteRef(fnId);
    return this.compiler.compileNode(fnRef, postAST);
  }
}

/**
 * Compiles an mdast footnote reference to hast, like:
 *
 *      Foo bar [^1] qux.
 *
 *      [^1]: Footnote definition.
 *
 * https://github.com/syntax-tree/mdast#footnotereference
 */
export class FootnoteReferenceCompiler implements MdastNodeCompiler {
  private constructor() {}

  static create(): FootnoteReferenceCompiler {
    return new FootnoteReferenceCompiler();
  }

  compileNode(node: unist.Node, _postAST: PostAST): unist.Node {
    md.checkType(node, 'footnoteReference', md.isFootnoteReference);
    const fnId = node.identifier;
    // There's also node.label which mdast defines as the original value of
    // the normalized node.identifier field.  We'll only use node.identifier
    // since I'm not sure when the label would ever be different than the ID.
    return FootnoteReferenceCompiler.makeHastNode(fnId);
  }

  static makeHastNode(fnId: string) {
    return h.elemProps('sup', { id: `fn-ref-${fnId}` }, [
      h.elemProps('a', { href: `#fn-${fnId}`, className: ['fn-ref'] }, [
        h.text(fnId),
      ]),
    ]);
  }
}

/**
 * Compiles an mdast heading to hast, like:
 *
 *     # Alpha
 *
 * https://github.com/syntax-tree/mdast#heading
 */
export class HeadingCompiler implements MdastNodeCompiler {
  private constructor(private readonly compiler: MdastCompiler) {}

  static create(compiler: MdastCompiler): HeadingCompiler {
    return new HeadingCompiler(compiler);
  }

  compileNode(node: unist.Node, postAST: PostAST): unist.Node {
    md.checkType(node, 'heading', md.isHeading);
    const children = this.compiler.compileChildren(node, postAST);
    return h.elem('h' + node.depth, children);
  }
}

/**
 * Compiles an mdast html node to hast, like:
 *
 *     <div></div>
 *
 * https://github.com/syntax-tree/mdast#html
 */
export class HTMLCompiler implements MdastNodeCompiler {
  private constructor() {}

  static create(): HTMLCompiler {
    return new HTMLCompiler();
  }

  compileNode(node: unist.Node, _postAST: PostAST): unist.Node {
    md.checkType(node, 'html', md.isHTML);
    return h.raw(node.value);
  }
}

/**
 * Compiles an mdast image node to hast, like:
 *
 *     ![alpha](https://example.com/favicon.ico "bravo")
 *
 * https://github.com/syntax-tree/mdast#image
 */
export class ImageCompiler implements MdastNodeCompiler {
  private constructor() {}

  static create(): ImageCompiler {
    return new ImageCompiler();
  }

  compileNode(node: unist.Node, _postAST: PostAST): unist.Node {
    md.checkType(node, 'image', md.isImage);
    const props: { src: string; alt?: string; title?: string } = {
      src: node.url,
    };
    if (node.alt) {
      props.alt = node.alt;
    }
    if (node.title) {
      props.title = node.title;
    }
    return h.elemProps('img', props);
  }
}

/**
 * Compiles an mdast imageReference node to hast, like:
 *
 *     ![alpha][bravo]
 *
 * https://github.com/syntax-tree/mdast#imagereference
 */
export class ImageReferenceCompiler implements MdastNodeCompiler {
  private constructor(private readonly compiler: MdastCompiler) {}

  static create(compiler: MdastCompiler): ImageReferenceCompiler {
    return new ImageReferenceCompiler(compiler);
  }

  compileNode(node: unist.Node, postAST: PostAST): unist.Node {
    md.checkType(node, 'imageReference', md.isImageRef);
    const id = node.identifier;
    let def = postAST.defsById.get(id);
    if (def === undefined) {
      return h.danglingImageRef(node);
    }
    const src = encodeURI(def.url.trim());
    const img = md.imageProps(src, { title: def.title, alt: node.alt });
    return this.compiler.compileNode(img, postAST);
  }
}

/**
 * Compiles an mdast inline code block to hast, like:
 *
 *     Foo bar `let a = 2;`.
 *
 * https://github.com/syntax-tree/mdast#inlinecode
 */
export class InlineCodeCompiler implements MdastNodeCompiler {
  private constructor() {}

  static create(): InlineCodeCompiler {
    return new InlineCodeCompiler();
  }

  compileNode(node: unist.Node, _postAST: PostAST): unist.Node {
    md.checkType(node, 'inline code', md.isInlineCode);
    return h.elemText('code', node.value);
  }
}

/**
 * Compiles an mdast inline code block to hast, like:
 *
 *     Foo bar `let a = 2;`.
 *
 * https://github.com/syntax-tree/mdast#inlinecode
 */
export class LinkCompiler implements MdastNodeCompiler {
  private constructor(private readonly compiler: MdastCompiler) {}

  static create(compiler: MdastCompiler): LinkCompiler {
    return new LinkCompiler(compiler);
  }

  compileNode(node: unist.Node, postAST: PostAST): unist.Node {
    md.checkType(node, 'link', md.isLink);
    const props: Partial<mdast.Link> = { href: encodeURI(node.url.trim()) };
    if (node.title) {
      props.title = node.title;
    }
    const children = this.compiler.compileChildren(node, postAST);
    return h.elemProps('a', props, children);
  }
}

/**
 * Compiles an mdast paragraph to hast, like:
 *
 *     Foo bar.
 *
 * https://github.com/syntax-tree/mdast#paragraph
 */
export class ParagraphCompiler implements MdastNodeCompiler {
  private constructor(private readonly compiler: MdastCompiler) {}

  static create(compiler: MdastCompiler): ParagraphCompiler {
    return new ParagraphCompiler(compiler);
  }

  compileNode(node: unist.Node, postAST: PostAST): unist.Node {
    md.checkType(node, 'paragraph', md.isParagraph);
    const children = this.compiler.compileChildren(node, postAST);
    return h.elem('p', children);
  }
}

/**
 * Compiles an mdast root to hast.
 *
 * https://github.com/syntax-tree/mdast#root
 */
export class RootCompiler implements MdastNodeCompiler {
  private constructor(private readonly compiler: MdastCompiler) {}

  static create(compiler: MdastCompiler): RootCompiler {
    return new RootCompiler(compiler);
  }

  compileNode(node: mdast.Root, postAST: PostAST): unist.Node {
    md.checkType(node, 'root', md.isRoot);
    const children = this.compiler.compileChildren(node, postAST);
    return h.elem('body', children);
  }
}

/**
 * Compiles an mdast strong block to hast, like:
 *
 *     This is **strong** and so is __this__.
 *
 * https://github.com/syntax-tree/mdast#strong
 */
export class StrongCompiler implements MdastNodeCompiler {
  private constructor(private readonly compiler: MdastCompiler) {}

  static create(compiler: MdastCompiler): StrongCompiler {
    return new StrongCompiler(compiler);
  }

  compileNode(node: unist.Node, postAST: PostAST): unist.Node {
    md.checkType(node, 'strong', md.isStrong);
    const children = this.compiler.compileChildren(node, postAST);
    return h.elem('strong', children);
  }
}

/**
 * Compiles a literal mdast text to hast.
 *
 * https://github.com/syntax-tree/mdast#text
 */
export class TextCompiler implements MdastNodeCompiler {
  private constructor() {}

  static create(): TextCompiler {
    return new TextCompiler();
  }

  compileNode(node: unist.Node, _postAST: PostAST): unist.Node {
    md.checkType(node, 'text', md.isText);
    return h.text(node.value);
  }
}

/**
 * Compiles a literal mdast text to hast.
 *
 * https://github.com/syntax-tree/mdast#text
 */
export class TomlCompiler implements MdastNodeCompiler {
  private constructor() {}

  static create(): TomlCompiler {
    return new TomlCompiler();
  }

  compileNode(node: unist.Node, _postAST: PostAST): unist.Node {
    md.checkType(node, 'toml', md.isToml);
    return unistNodes.ignored();
  }
}
