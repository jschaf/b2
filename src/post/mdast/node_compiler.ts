import { checkDefined, checkState } from '//asserts';
import { MdastCompiler } from '//post/mdast/compiler';
import { hastElem, hastElemWithProps, hastText } from '//post/hast/hast_nodes';
import * as md from '//post/mdast/nodes';
import { PostAST } from '//post/post_ast';
import { isString } from '//strings';
import * as mdast from 'mdast';
import * as unist from 'unist';

/** Compiler for a single mdast node. */
export interface MdastNodeCompiler {
  compileNode(node: unist.Node, ast: PostAST): unist.Node;
}

/**
 * Compiles an mdast blockquote to hast, like:
 *
 *     > foo bar
 *     > baz qux
 */
export class BlockquoteCompiler implements MdastNodeCompiler {
  private constructor(private readonly compiler: MdastCompiler) {}

  static create(compiler: MdastCompiler): BlockquoteCompiler {
    return new BlockquoteCompiler(compiler);
  }

  compileNode(node: unist.Node, postAST: PostAST): unist.Node {
    md.checkType(node, 'blockquote', md.isBlockquote);
    const children = this.compiler.compileChildren(node, postAST);
    return hastElem('blockquote', children);
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
    return hastElem('break');
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
    return hastElem('pre', [
      hastElemWithProps('code', props, [hastText(node.value)]),
    ]);
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
    return hastElem('del', children);
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
    return hastElem('em', children);
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
    return hastElemWithProps('sup', { id: `fn-ref-${fnId}` }, [
      hastElemWithProps('a', { href: `#fn-${fnId}`, className: ['fn-ref'] }, [
        hastText(fnId),
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
    return hastElem('h' + node.depth, children);
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
    return hastElem('p', children);
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
    return hastElem('body', children);
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
    return hastText(node.value);
  }
}
