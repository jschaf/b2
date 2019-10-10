import Koa from 'koa';

import KoaRouter from 'koa-router';
import koaBody from 'koa-body';
import MdIt from 'markdown-it';
import * as fs from 'fs';
import * as dates from './dates';
import * as strings from './strings';
import * as zipFiles from "./zip_files";
import flags from 'flags';
import git from "nodegit";
import Token from 'markdown-it/lib/token';
import yaml from 'js-yaml';

const gitDirFlag = flags.defineString('git-dir').setDescription('The path to the git dir.');
flags.parse();

const app = new Koa();

const router = new KoaRouter();
const BUNDLE_PREFIX = 'Content.textbundle/';

console.log('!!! gitFlag', gitDirFlag.currentValue);

// @ts-ignore
const getCommitMessage = async (repoDir: string): Promise<string> => {
  const repo: git.Repository = await git.Repository.open(repoDir);
  const commit = await repo.getBranchCommit('master');
  return commit.message();
};

getCommitMessage(gitDirFlag.currentValue).then(msg => console.log('MSG1' + msg)).catch(err => console.log('MSG: ' + err));


type Schema = Record<string, { type: 'string' | 'Date', isRequired: boolean }>;
const METADATA_SCHEMA: Schema = {
  slug: {type: 'string', isRequired: true},
  date: {type: 'Date', isRequired: true},
  publish_state: {type: 'string', isRequired: false},
};


const checkMetadataSchema = (metadata: any): typeof METADATA_SCHEMA => {
  for (const [key, {isRequired}] of Object.entries(METADATA_SCHEMA)) {
    if (isRequired && !metadata.hasOwnProperty(key)) {
      throw new Error(`YAML metadata missing required key ${key}.`);
    }
  }

  for (const [key, value] of Object.entries(metadata)) {
    if (!METADATA_SCHEMA.hasOwnProperty(key)) {
      throw new Error((`Extra property key ${key} in YAML.`));
    }
    const schemaDef = METADATA_SCHEMA[key];
    switch (schemaDef.type) {
      case 'Date':
        if (!dates.isValidDate(value)) {
          throw new Error(`Invalid date: ${value} for key: ${key}.`);
        }
        break;
      case 'string':
        if (!strings.isString(value)) {
          throw new Error(`Expected string for key ${key} but got ${value}.`);
        }
        break;
    }
  }
  return metadata;
};

const findMetadataIndex = (tokens: Token[]): number => {
  const maxTokensToSearch = 20;
  const index = tokens.findIndex(t => t.type === 'fence' && t.content.startsWith('# Metadata'));
  if (index === -1) {
    throw new Error(`Unable to find a YAML metadata section in `
        + `the first ${maxTokensToSearch} tokens.`)
  }
  return index;
};

const parseMetadata = (token: Token): Schema => {
  const rawYaml = yaml.safeLoad(token.content);
  return checkMetadataSchema(rawYaml);
};

const doThing = async (path: string): Promise<void> => {
  const compressed = await fs.promises.readFile(path);
  const zipFile = await zipFiles.unzipFromBuffer(compressed);
  const entries = await zipFiles.readAllEntries(zipFile);
  console.log('!!! files', JSON.stringify(entries.map(e => e.filePath)));

  const texts = entries.filter(e => e.filePath === BUNDLE_PREFIX + 'text.md');
  if (texts.length !== 1) {
    throw new Error('Unable to find text.md in entries: '
        + entries.map(e => e.filePath));
  }
  const text = texts[0];
  console.log('!!! text', text.contents.toString('utf8'));
  const md = new MdIt();
  const tokens = md.parse(text.contents.toString('utf8'), {});
  const index = findMetadataIndex(tokens);
  const metadata = parseMetadata(tokens[index]);
  tokens.splice(index, 1);
  md.re
};

doThing('/Users/joe/gorilla.textpack').finally(() => console.log('done'));

router.get('/', async (ctx) => {
  ctx.response.body = 'hello, world';
});

/**
 * Receives a file from a multi-part form upload and commits the markdown file
 * into the Git repo.
 */
router.post('/commit_post', koaBody({multipart: true}), async (ctx) => {
  // console.log('!!! Got POST request', ctx);
  ctx.body = 'hello world';


  // if (ctx.request.files == null) {
  //   throw new Error("No files in request");
  // }
  // const file = ctx.request.files.file;
  // const compressed = await fs.promises.readFile(file.path);
  // const zipFile = await zipFiles.unzipFromBuffer(compressed);
  // const entries = await zipFiles.readAllEntries(zipFile);
  // console.log('!!! files', JSON.stringify(entries.map(e => e.filePath)));
  //
  // const texts = entries.filter(e => e.filePath === BUNDLE_PREFIX + 'text.md');
  // if (texts.length !== 1) {
  //   throw new Error('Unable to find text.md in entries: '
  //       + entries.map(e => e.filePath));
  // }
  // const text = texts[0];
  // console.log('!!! text', text.contents.toString('utf8'));
  // const md = new MarkdownIt();
  // const tokens = md.parse(text.contents.toString('utf8'), {});
  // console.log('!!! tokens', tokens.slice(5));

  // const html = md.render(text.contents.toString('utf8'));
  // console.log('!!! html', html);

  // Format markdown
  // Get slug from metadata
  // Check that git dir is clean
  // Overwrite file posts/$SLUG.md
  //

  // Decompress

  // Compile markdown to HTML
  // - katex
  // - citations
  // - compile markdown

});

app.use(router.routes());
// app.use(router.allowedMethods());
app.listen(3000);
console.log('Server started on port 3000');
