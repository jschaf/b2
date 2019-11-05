import flags from 'flags';
import * as fs from 'fs';
import Koa from 'koa';
import koaBody from 'koa-body';

import KoaRouter from 'koa-router';
import { PostBag } from './post/post_bag';
import { PostCommitter } from './post/post_committer';

const gitDirFlag = flags
  .defineString('git-dir')
  .setDescription('The path to the git dir.');
flags.parse();

const app = new Koa();

const router = new KoaRouter();

console.log('!!! gitFlag', gitDirFlag.currentValue);

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

  if (ctx.request.files == null) {
    throw new Error('No files in request');
  }
  const file = ctx.request.files.file;
  const textPack = await fs.promises.readFile(file.path);
  const bag = await PostBag.fromTextPack(textPack);
  const committer = PostCommitter.forFs(fs, gitDirFlag.currentValue);
  await committer.commit(bag);
  //  await committer.pushOrigin();

  // Compile markdown to HTML
  // - katex
  // - citations
  // - compile markdown
});

app.use(router.routes());
// app.use(router.allowedMethods());
app.listen(3000);
console.log('Server started on port 3000');
