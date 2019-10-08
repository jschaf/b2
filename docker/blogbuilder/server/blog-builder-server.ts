import Koa from 'koa';

import KoaRouter from 'koa-router';
import koaBody from 'koa-body';
import * as fs from 'fs';
import * as zipFiles from "./zip_files";

const app = new Koa();

const router = new KoaRouter();

router.get('/', async (ctx) => {
  ctx.response.body = 'hello, world';
});

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
  const files = await zipFiles.readFilesToMap(zipFile);
  console.log('!!! files', JSON.stringify(files));

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
