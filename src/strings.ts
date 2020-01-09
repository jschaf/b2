/**
 * Returns true if value is a string.
 *
 * See https://stackoverflow.com/a/9436948/30900.
 */
export const isString = (value: unknown): value is string => {
  return typeof value === 'string' || value instanceof String;
};

/**
 * Returns true if value is a string or undefined (but not null).
 */
export const isOptionalString = (
  value: unknown
): value is string | undefined => {
  return value === undefined || isString(value);
};

/**
 * A tagged template that remove leading indentation from tagged template lines.
 *
 * For example:
 *
 *     const foo = dedent`
 *       foo
 *         bar
 *     `;
 *     expect(foo).toEqual('foo\n  bar');
 *
 */
export const dedent = (
  literals: TemplateStringsArray,
  ...placeholders: string[]
): string => {
  let raw = '';
  for (let i = 0; i < literals.length; i++) {
    raw += literals[i]
      // Join lines when there is a suppressed newline.
      .replace(/\\\n[ \t]*/g, '')
      // Handle escaped backticks.
      .replace(/\\`/g, '`');

    if (i < placeholders.length) {
      raw += placeholders[i];
    }
  }
  // Calculate min line width to chop from each line.
  const lines = raw.split('\n');
  let minIndent: number | null = null;
  for (const line of lines) {
    const m = /^(\s+)\S+/.exec(line);
    if (m) {
      const indent = m[1].length;
      if (minIndent === null) {
        minIndent = indent;
      } else {
        minIndent = Math.min(indent, minIndent);
      }
    }
  }

  // Chop min indent width from each line.
  let result: string;
  if (minIndent !== null) {
    result = lines
      .map(l => (l.startsWith(' ') ? l.slice(minIndent || 0) : l))
      .join('\n');
  } else {
    result = raw;
  }

  return result.trim();
};

export class StringBuilder {
  private buf: Buffer;
  private usedLength: number = 0;

  private constructor() {
    this.buf = Buffer.allocUnsafe(16);
  }

  static create(): StringBuilder {
    return new StringBuilder();
  }

  writeString(s: string): void {
    const remaining = this.buf.length - this.usedLength;
    if (remaining < s.length) {
      this.reallocate(this.buf.length + s.length);
    }
    this.buf.write(s, this.usedLength, 'utf8');
    this.usedLength += s.length;
  }

  toString(): string {
    const start = 0;
    const end = this.usedLength;
    return this.buf.toString('utf8', start, end);
  }

  private reallocate(min: number) {
    let newLen = this.buf.length * 2;
    while (newLen < min) {
      newLen *= 2;
    }
    const target = Buffer.allocUnsafe(newLen);
    const targetStart = 0;
    const sourceStart = 0;
    const sourceEnd = this.usedLength;
    this.buf.copy(target, targetStart, sourceStart, sourceEnd);
    this.buf = target;
  }
}
