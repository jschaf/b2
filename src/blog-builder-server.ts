import Koa from 'koa';

import KoaRouter from 'koa-router';
import koaBody from 'koa-body';
// import MarkdownIt from 'markdown-it';
import * as fs from 'fs';
import * as zipFiles from "./zip_files";

const app = new Koa();

const router = new KoaRouter();
const BUNDLE_PREFIX = 'Content.textbundle/';

router.get('/', async (ctx) => {
  ctx.response.body = 'hello, world';
});

/**
 * Receives a file from a multi-part form upload and commits the markdown file
 * into the Git repo.
 */
router.post('/commit_post', koaBody({multipart: true}), async (ctx) => {
  console.log('!!! Got POST request', ctx);
  ctx.body = 'hello world';


  console.log('!!! files', ctx.request.files);
  if (ctx.request.files == null) {
    throw new Error("No files in request");
  }
  const file = ctx.request.files.file;
  const compressed = await fs.promises.readFile(file.path);
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
  // const md = new MarkdownIt();
  // const html = md.render(text.contents.toString('utf8'));
  // console.log('!!! html', html);

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
