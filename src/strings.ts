/**
 * Returns true if value is a string.
 *
 * See https://stackoverflow.com/a/9436948/30900.
 */
export const isString = (value: any): boolean => {
  return typeof value === 'string' || value instanceof String;
};

/**
 * Removes leading indentation from tagged template lines.
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
    let m = line.match(/^(\s+)\S+/);
    if (m) {
      let indent = m[1].length;
      if (minIndent === null) {
        minIndent = indent;
      } else {
        minIndent = Math.min(indent, minIndent);
      }
    }
  }

  // Chop min indent width from each line.
  let result = '';
  if (minIndent !== null) {
    result = lines
      .map(l => (l.charAt(0) === ' ' ? l.slice(minIndent || 0) : l))
      .join('\n');
  } else {
    result = raw;
  }

  return result.trim();
};
