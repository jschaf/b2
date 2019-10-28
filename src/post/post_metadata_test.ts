import { PostMetadata } from './post_metadata';
import * as dates from '../dates';
import { dedent } from '../strings';
import unified from 'unified';
import remarkParse from 'remark-parse';

const processor = unified().use(remarkParse);

test('parses valid tokens', () => {
  const tree = processor.parse(dedent`
    # hello
    
    \`\`\`yaml
    # Metadata
    slug: foo_bar
    date: 2019-10-08
    \`\`\`
  `);

  const metadata = PostMetadata.parseFromMarkdownAST(tree);

  expect(metadata.schema).toEqual({
    date: dates.fromISO('2019-10-08'),
    slug: 'foo_bar',
  });
});
