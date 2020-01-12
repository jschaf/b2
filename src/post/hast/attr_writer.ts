import { checkArg } from '//asserts';
import { HTMLEscaper } from '//post/html/escapers';
import { StringBuilder } from '//strings';

const validNameChars = /^[-_a-zA-Z0-9]+$/;

/**
 * AttrWriter normalizes HTML attributes and writes them into a string builder.
 */
export class AttrWriter {
  private constructor() {}

  static create(): AttrWriter {
    return new AttrWriter();
  }

  writeElemProps(props: Record<string, unknown>, sb: StringBuilder): void {
    const entries = Object.entries(props);
    if (entries.length === 0) {
      return;
    }
    for (let i = 0; i < entries.length; i++) {
      const [key, value] = entries[i];
      this.writeAttr(key, value, sb);
      if (i < entries.length - 1) {
        sb.writeString(' ');
      }
    }
  }

  private writeAttr(name: string, value: unknown, sb: StringBuilder): void {
    const n = normalizeName(name);
    const v = normalizeValue(value);
    if (n === null || v === null) {
      return;
    }

    sb.writeString(n);
    if (v !== '') {
      sb.writeString('="');
      sb.writeString(v);
      sb.writeString('"');
    }
  }
}

/**
 * normalizeName returns null if name is an invalid HTML attribute name,
 * otherwise returns the string.
 *
 * https://html.spec.whatwg.org/multipage/syntax.html#syntax-attribute-name
 */
const normalizeName = (name: string): string | null => {
  if (name.length === 0) {
    return null;
  }
  if (!validNameChars.test(name)) {
    return null;
  }
  return name;
};

const escaper = HTMLEscaper.create();

/**
 * normalizeValue transforms a string so it's a valid HTML attribute value.
 * Returns null if the entire attributed should be omitted.
 *
 * https://html.spec.whatwg.org/multipage/syntax.html#syntax-attribute-value
 */
const normalizeValue = (value: unknown): string | null => {
  switch (typeof value) {
    case 'boolean':
      if (!value) {
        return null;
      }
      return '';

    case 'number':
      checkArg(!isNaN(value), 'expected valid number, got NaN');
      return value.toString();

    case 'bigint':
      return value.toString();

    case 'string':
      if (escaper.needsEscaped(value)) {
        return escaper.escape(value);
      }

      return value;

    case 'object':
      if (value === null) {
        return '';
      } else if (Array.isArray(value)) {
        return value.map(v => normalizeValue(v)).join(' ');
      } else {
        // Other types include regexps and dates.
        return '';
      }

    case 'function':
    case 'symbol':
    case 'undefined':
      return '';

    default:
      throw new Error('unreachable');
  }
};
