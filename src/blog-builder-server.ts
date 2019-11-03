import Koa from 'koa';

import KoaRouter from 'koa-router';
import koaBody from 'koa-body';
import * as fs from 'fs';
import flags from 'flags';
import git from 'nodegit';
import { PostParser } from './post/post_parser';
import {Unzipper} from "./zip_files";

const gitDirFlag = flags
  .defineString('git-dir')
  .setDescription('The path to the git dir.');
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

getCommitMessage(gitDirFlag.currentValue)
  .then(msg => console.log('MSG1' + msg))
  .catch(err => console.log('MSG: ' + err));

const doThing = async (path: string): Promise<void> => {
  const compressed = await fs.promises.readFile(path);
  const entries = await Unzipper.unzip(compressed);
  console.log('!!! files', JSON.stringify(entries.map(e => e.filePath)));

  const texts = entries.filter(e => e.filePath === BUNDLE_PREFIX + 'text.md');
  if (texts.length !== 1) {
    throw new Error(
      'Unable to find text.md in entries: ' + entries.map(e => e.filePath)
    );
  }
  const text = texts[0];
  console.log('!!! text', text.contents.toString('utf8'));
  const postParser = PostParser.create();
  const postNode = postParser.parseMarkdown(text.contents.toString('utf8'));
  console.log('!!! postNode', postNode);
};

doThing('/Users/joe/gorilla.textpack').finally(() => console.log('done'));

router.get('/', async ctx => {
  ctx.response.body = 'hello, world';
});

/**
 * Receives a file from a multi-part form upload and commits the markdown file
 * into the Git repo.
 */
router.post('/commit_post', koaBody({ multipart: true }), async ctx => {
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
