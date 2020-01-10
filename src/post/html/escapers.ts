/**
 * HTMLEscaper escapes HTML.
 *
 * Based on Guava:
 * https://guava.dev/releases/19.0/api/docs/src-html/com/google/common/html/HtmlEscapers.html#line.61
 */
export class HTMLEscaper {
  private constructor() {}

  static create(): HTMLEscaper {
    return new HTMLEscaper();
  }

  escape(s: string): string {
    return (
      s
        // An ambiguous ampersand is an (&) that is followed by one or more
        // ASCII alphanumerics, followed by a semicolon character.
        .replace(/&([a-zA-Z0-9]+;)/g, '&amp;$1')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
    );
  }

  needsEscaped(s: string): boolean {
    return /((&[a-zA-Z0-9]+;)|["'<>])/.test(s);
  }
}
