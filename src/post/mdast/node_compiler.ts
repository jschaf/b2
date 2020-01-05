import { checkDefined, checkState } from '//asserts';
import { isBoolean } from '//booleans';
import * as h from '//post/hast/nodes';
import { MdastCompiler } from '//post/mdast/compiler';
import * as md from '//post/mdast/nodes';
import { isNumber } from '//post/numbers';
import { PostAST } from '//post/post_ast';
import { isString } from '//strings';
import * as mdast from 'mdast';
import * as hast from 'hast-format';
import * as unist from 'unist';

/** Compiler for a single mdast node. */
export interface MdastNodeCompiler {
  compileNode(node: unist.Node, ast: PostAST): unist.Node[];
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

  compileNode(node: unist.Node, postAST: PostAST): unist.Node[] {
    md.checkType(node, 'blockquote', md.isBlockquote);
    const children = this.compiler.compileChildren(node, postAST);
    return [h.elem('blockquote', children)];
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

  compileNode(node: unist.Node, _postAST: PostAST): unist.Node[] {
    md.checkType(node, 'break', md.isBreak);
    return [h.elem('break')];
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

  compileNode(node: unist.Node, _postAST: PostAST): unist.Node[] {
    md.checkType(node, 'code', md.isCode);
    const props: Record<string, unknown> = {};
    if (isString(node.lang) && node.lang !== '') {
      props.className = ['lang-' + node.lang.trim()];
    }
    return [h.elem('pre', [h.elemProps('code', props, [h.text(node.value)])])];
  }
}

/**
 * Compiles deleted mdast text to hast, like:
 *
 *     Foo bar ~~this is deleted~~.
 *
 * https://github.com/syntax-tree/mdast#delete
 */
export class DeleteCompiler implements MdastNodeCompiler {
  private constructor(private readonly compiler: MdastCompiler) {}

  static create(compiler: MdastCompiler): DeleteCompiler {
    return new DeleteCompiler(compiler);
  }

  compileNode(node: unist.Node, postAST: PostAST): unist.Node[] {
    md.checkType(node, 'delete', md.isDelete);
    const children = this.compiler.compileChildren(node, postAST);
    return [h.elem('del', children)];
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

  compileNode(node: unist.Node, postAST: PostAST): unist.Node[] {
    md.checkType(node, 'emphasis', md.isEmphasis);
    const children = this.compiler.compileChildren(node, postAST);
    return [h.elem('em', children)];
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

  compileNode(node: unist.Node, postAST: PostAST): unist.Node[] {
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

  compileNode(node: unist.Node, _postAST: PostAST): unist.Node[] {
    md.checkType(node, 'footnoteReference', md.isFootnoteReference);
    const fnId = node.identifier;
    // There's also node.label which mdast defines as the original value of
    // the normalized node.identifier field. Using identifier since it's
    // normalized.
    return [FootnoteReferenceCompiler.makeHastNode(fnId)];
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

  compileNode(node: unist.Node, postAST: PostAST): unist.Node[] {
    md.checkType(node, 'heading', md.isHeading);
    const children = this.compiler.compileChildren(node, postAST);
    return [h.elem('h' + node.depth, children)];
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

  compileNode(node: unist.Node, _postAST: PostAST): unist.Node[] {
    md.checkType(node, 'html', md.isHTML);
    return [h.raw(node.value)];
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

  compileNode(node: unist.Node, _postAST: PostAST): unist.Node[] {
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
    return [h.elemProps('img', props)];
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

  compileNode(node: unist.Node, postAST: PostAST): unist.Node[] {
    md.checkType(node, 'imageReference', md.isImageRef);
    let def = postAST.getDefinition(node.identifier);
    if (def === null) {
      return [h.danglingImageRef(node)];
    }
    const src = h.normalizeUri(def.url);
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

  compileNode(node: unist.Node, _postAST: PostAST): unist.Node[] {
    md.checkType(node, 'inline code', md.isInlineCode);
    return [h.elemText('code', node.value)];
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

  compileNode(node: unist.Node, postAST: PostAST): unist.Node[] {
    md.checkType(node, 'link', md.isLink);
    const props: Partial<mdast.Link> = { href: h.normalizeUri(node.url) };
    if (node.title) {
      props.title = node.title;
    }
    const children = this.compiler.compileChildren(node, postAST);
    return [h.elemProps('a', props, children)];
  }
}

/**
 * Compiles an mdast link reference to hast, like:
 *
 *     [alpha][bravo]
 *
 * https://github.com/syntax-tree/mdast#linkreference
 */
export class LinkReferenceCompiler implements MdastNodeCompiler {
  private constructor(private readonly compiler: MdastCompiler) {}

  static create(compiler: MdastCompiler): LinkReferenceCompiler {
    return new LinkReferenceCompiler(compiler);
  }

  compileNode(node: unist.Node, postAST: PostAST): unist.Node[] {
    md.checkType(node, 'linkReference', md.isLinkRef);
    const def = postAST.getDefinition(node.identifier);
    if (def === null) {
      const c = (n: mdast.LinkReference) =>
        this.compiler.compileChildren(n, postAST);
      return h.danglingLinkRef(node, c);
    }

    const props: md.LinkProps = {};
    if (def.title) {
      props.title = def.title;
    }
    const link = md.linkProps(def.url, props, node.children);
    return this.compiler.compileNode(link, postAST);
  }
}

/**
 * Compiles an mdast list to hast, like:
 *
 *     - foo
 *     - bar
 *
 * https://github.com/syntax-tree/mdast#list
 */
export class ListCompiler implements MdastNodeCompiler {
  private constructor(private readonly compiler: MdastCompiler) {}

  static CHECKBOX_CLASS_NAME = 'task-list';

  static create(compiler: MdastCompiler): ListCompiler {
    return new ListCompiler(compiler);
  }

  compileNode(node: unist.Node, postAST: PostAST): unist.Node[] {
    md.checkType(node, 'list', md.isList);
    const tag = node.ordered ? 'ol' : 'ul';
    const props: Record<string, unknown> = {};
    if (isNumber(node.start) && node.start !== 1) {
      props.start = node.start;
    }
    if (ListCompiler.hasCheckboxItem(node)) {
      props.className = [ListCompiler.CHECKBOX_CLASS_NAME];
    }
    const children = this.compiler.compileChildren(node, postAST);
    return [h.elemProps(tag, props, children)];
  }

  private static hasCheckboxItem(node: mdast.List): boolean {
    for (const child of node.children) {
      if (isBoolean(child.checked)) {
        return true;
      }
    }
    return false;
  }
}

enum ListItemCheckedState {
  Normal,
  Checked,
  Unchecked,
}

/**
 * Compiles an mdast list item to hast, like:
 *
 *     - foo bar
 *
 * https://github.com/syntax-tree/mdast#listitem
 */
export class ListItemCompiler implements MdastNodeCompiler {
  private constructor(private readonly compiler: MdastCompiler) {}

  static create(compiler: MdastCompiler): ListItemCompiler {
    return new ListItemCompiler(compiler);
  }

  compileNode(node: unist.Node, postAST: PostAST): unist.Node[] {
    md.checkType(node, 'listItem', md.isListItem);
    const children = this.compiler.compileChildren(node, postAST);
    const isLoose = node.spread === true;

    switch (ListItemCompiler.getCheckedState(node)) {
      case ListItemCheckedState.Normal:
        if (isLoose) {
          return [h.elem('li', children)];
        } else {
          const unwrapped = ListItemCompiler.unwrapParagraphs(children);
          return [h.elem('li', unwrapped)];
        }

      case ListItemCheckedState.Checked:
      case ListItemCheckedState.Unchecked:
        checkState(isBoolean(node.checked));
        const checkbox = ListItemCompiler.checkbox(node.checked);
        if (isLoose) {
          return [h.elem('li', [checkbox, ...children])];
        } else {
          const unwrapped = ListItemCompiler.unwrapParagraphs(children);
          return [h.elem('li', [checkbox, ...unwrapped])];
        }
    }
  }

  static checkbox(checked: boolean): hast.Element {
    return h.elemProps('input', { type: 'checkbox', checked, disabled: true });
  }

  private static unwrapParagraphs(children: unist.Node[]): unist.Node[] {
    const rs = [];
    for (const c of children) {
      if (h.isElem('p', c)) {
        rs.push(...c.children);
      } else {
        rs.push(c);
      }
    }
    return rs;
  }

  private static getCheckedState(n: mdast.ListItem): ListItemCheckedState {
    if (n.checked === null || n.checked === undefined) {
      return ListItemCheckedState.Normal;
    } else if (n.checked) {
      return ListItemCheckedState.Checked;
    } else {
      return ListItemCheckedState.Unchecked;
    }
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

  compileNode(node: unist.Node, postAST: PostAST): unist.Node[] {
    md.checkType(node, 'paragraph', md.isParagraph);
    const children = this.compiler.compileChildren(node, postAST);
    return [h.elem('p', children)];
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

  compileNode(node: mdast.Root, postAST: PostAST): unist.Node[] {
    md.checkType(node, 'root', md.isRoot);
    const children = this.compiler.compileChildren(node, postAST);
    return [h.elem('body', children)];
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

  compileNode(node: unist.Node, postAST: PostAST): unist.Node[] {
    md.checkType(node, 'strong', md.isStrong);
    const children = this.compiler.compileChildren(node, postAST);
    return [h.elem('strong', children)];
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

  compileNode(node: unist.Node, _postAST: PostAST): unist.Node[] {
    md.checkType(node, 'text', md.isText);
    return [h.text(node.value)];
  }
}

/**
 * Compiles an mdast thematicBreak to hast, like:
 *
 *     foo bar
 *
 *     ***
 *
 *     baz
 *
 * https://github.com/syntax-tree/mdast#thematicbreak
 */
export class ThematicBreakCompiler implements MdastNodeCompiler {
  private constructor() {}

  static create(): ThematicBreakCompiler {
    return new ThematicBreakCompiler();
  }

  compileNode(node: unist.Node, _postAST: PostAST): unist.Node[] {
    md.checkType(node, 'thematicBreak', md.isThematicBreak);
    return [h.elem('hr')];
  }
}

/**
 * Compiles an mdast table to hast, like:
 *
 *     | foo | bar |
 *     | --- | --- |
 *     | baz | qux |
 *
 *  The table compiler also handles TableRow and TableCell because it's easier
 *  to determine whether to use <th> or <td>.
 *
 * https://github.com/syntax-tree/mdast#table
 */
export class TableCompiler implements MdastNodeCompiler {
  private constructor(private readonly compiler: MdastCompiler) {}

  static create(compiler: MdastCompiler): TableCompiler {
    return new TableCompiler(compiler);
  }

  compileNode(node: unist.Node, postAST: PostAST): unist.Node[] {
    md.checkType(node, 'table', md.isTable);
    const rows: hast.Element[] = [];
    for (const rowNode of node.children) {
      md.checkType(rowNode, 'table row', md.isTableRow);
      const cells: hast.Element[] = [];

      for (const cellNode of rowNode.children) {
        md.checkType(cellNode, 'table cell', md.isTableCell);
        const c = this.compiler.compileChildren(cellNode, postAST);
        cells.push(h.elem('td', c));
      }

      const tr = h.elem('tr', cells);
      rows.push(tr);
    }

    const head = h.elem('thead', [rows[0]]);
    const tableChildren = [head];
    if (rows.length > 1) {
      const body = h.elem('tbody', rows.slice(1));
      tableChildren.push(body);
    }

    return [h.elem('table', tableChildren)];
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

  compileNode(node: unist.Node, _postAST: PostAST): unist.Node[] {
    md.checkType(node, 'toml', md.isToml);
    return [];
  }
}
