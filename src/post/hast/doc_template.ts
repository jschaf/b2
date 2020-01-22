import { isParentTag } from '//post/hast/nodes';
import * as h from '//post/hast/nodes';
import { PostType } from '//post/metadata';
import * as hast from 'hast-format';
import * as unist from 'unist';

/** DocTemplate is an hast template for a full HTML document. */
export class DocTemplate {
  private readonly head: unist.Node[] = [];

  private constructor() {}

  static create(): DocTemplate {
    return new DocTemplate();
  }

  static templates(): Map<PostType, DocTemplate> {
    return new Map<PostType, DocTemplate>(templates);
  }

  addToHead(...hs: unist.Node[]): DocTemplate {
    this.head.push(...hs);
    return this;
  }

  render(children: unist.Node[]): hast.Root {
    let nodes = children;
    if (children.length === 1) {
      const body = children[0];
      if (isParentTag('body', body)) {
        nodes = body.children;
      }
    }

    return h.root([
      h.doctype(),
      h.elemProps('html', { lang: 'en' }, [
        h.elem('head', this.head),
        h.elem('body', [
          h.elem('header'),
          h.elem('main', [
            h.elemProps('div', { className: ['main-inner-container'] }, nodes),
          ]),
          h.elem('footer'),
        ]),
      ]),
    ]);
  }
}

const postTemplate = () => {
  return DocTemplate.create().addToHead(
    // Sets the encoding when not present in the Content-Type header
    // https://stackoverflow.com/a/16506858/30900
    h.elemProps('meta', { charset: 'utf-8' }),

    // Skipping http-equiv because we don't need to support IE8 or IE9.
    // <meta http-equiv=x-ua-compatible content="IE=edge,chrome=1">
    // https://stackoverflow.com/a/6771584/30900

    // Make the site full width on mobile.
    // https://stackoverflow.com/a/16532471/30900
    h.elemProps('meta', {
      name: 'viewport',
      content: 'width=device-width, initial-scale=1.0',
    }),

    // Allow spiders to crawl and index the site.
    // https://stackoverflow.com/a/51277688/30900
    h.elemProps('meta', { name: 'robots', content: 'index, follow' }),

    h.elemProps('link', { rel: 'icon', href: '/favicon.ico' }),
    h.elemProps('link', {
      rel: 'apple-touch-icon-precomposed',
      href: '/favicon-152.png',
    }),

    h.elemProps('script', {
      defer: true,
      src: '/instantpage.min.js',
      type: 'application/javascript',
    })
  );
};

const templates: [PostType, DocTemplate][] = [[PostType.Post, postTemplate()]];
