import MdIt from 'markdown-it';
import {Metadata} from "./metadata";
import * as dates from '../dates';


test('parses valid tokens', () => {
  const md = new MdIt();
  const tokens = md.parse('```\n# Metadata\nslug: foo_bar\ndate: 2019-10-08\n```', {});

  const metadata = Metadata.parseFromMarkdownTokens(tokens);

  expect(metadata.schema).toEqual({
    date: dates.fromISO('2019-10-08'),
    slug: 'foo_bar',
  });
});

test('parses valid tokens with preceding tokens', () => {
  const md = new MdIt();
  const tokens = md.parse([
      '# Title',
      '',
      'Some text',
      '',
      '```',
      '# Metadata',
      'slug: foo_bar',
      'date: 2019-10-08',
      '```'
      ].join('\n'), {});

  const metadata = Metadata.parseFromMarkdownTokens(tokens);

  expect(metadata.schema).toEqual({
    date: dates.fromISO('2019-10-08'),
    slug: 'foo_bar',
  });
});
