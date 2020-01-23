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
      const root = children[0];
      if (isParentTag('root', root) || isParentTag('body', root)) {
        nodes = root.children;
      }
    }

    const siteNav = h.elemProps(
      'nav',
      { className: ['site-nav'], role: 'navigation' },
      [
        h.elemProps(
          'a',
          {
            className: ['site-title'],
            href: '/',
            title: 'Home page',
          },
          [h.text('Joe Schafer')]
        ),
        h.elem('ul', [
          h.elem('li', [
            h.elemProps(
              'a',
              { href: 'https://github.com/jschaf', title: 'GitHub page' },
              [h.text('GitHub')]
            ),
          ]),
          h.elem('li', [
            h.elemProps(
              'a',
              {
                href: 'https://www.linkedin.com/in/jschaf/',
                title: 'LinkedIn page',
              },
              [h.text('LinkedIn')]
            ),
          ]),
        ]),
      ]
    );
    const currentYear = new Date().getFullYear();
    // https://developer.mozilla.org/en-US/docs/Web/Accessibility/ARIA/Roles/Contentinfo_role
    const footer = h.elemProps('footer', { role: 'contentinfo' }, [
      h.elemProps('a', { href: '/', title: 'Home page' }, [
        h.text(`© ${currentYear} Joe Schafer`),
      ]),
    ]);
    return h.root([
      h.doctype(),
      h.elemProps('html', { lang: 'en' }, [
        h.elem('head', this.head),
        h.elem('body', [
          h.elem('header', [siteNav]),
          h.elem('main', [
            h.elemProps('div', { className: ['main-inner-container'] }, nodes),
          ]),
          footer,
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

    h.elemProps('link', { rel: 'stylesheet', href: '/style/main.css' }),
    h.elemProps('script', {
      defer: true,
      src: '/instantpage.min.js',
      type: 'application/javascript',
    })
  );
};

const templates: [PostType, DocTemplate][] = [[PostType.Post, postTemplate()]];
