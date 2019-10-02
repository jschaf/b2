---
title: Faster Git checkout for continuous integration
subtitle: Reduce `git checkout` from 30 seconds to 2 seconds on CircleCI.
date: 2019-09-20

---
The default CircleCI `checkout` step is slow because it downloads the
entire Git repository history in two remote fetches. My company’s
source repository with 7 years of history is 500MB and takes 30
seconds to `git clone` on CircleCI. Using a shallow checkout and downloading the repository at a specific hash in a single remote fetch reduces the checkout step time to 2 seconds.

## The default CircleCI `checkout` step

We’ll start by analyzing the builtin CircleCI `checkout` step. The
interesting parts of the default [CircleCI ](https://circleci.com/docs/2.0/configuration-reference/#checkout)`[checkout](https://circleci.com/docs/2.0/configuration-reference/#checkout)`
step are below and the full code is available at this [GitHub Gist](https://gist.github.com/jschaf/31d88678cbf733e9bb749ec0afdcc418).

```bash
#!/bin/bash

# The default CircleCI checkout step.

git clone "$CIRCLE_REPOSITORY_URL" .

if [ -n "$CIRCLE_TAG" ]; then
  git fetch --force origin "refs/tags/${CIRCLE_TAG}"
else
  git fetch --force origin ":remotes/origin/"
fi

if [ -n "$CIRCLE_TAG" ]; then
  git reset --hard "$CIRCLE_SHA1"
  git checkout -q "$CIRCLE_TAG"
elif [ -n "$CIRCLE_BRANCH" ]; then
  git reset --hard "$CIRCLE_SHA1"
  git checkout -q -B "$CIRCLE_BRANCH"
fi

git reset --hard "$CIRCLE_SHA1"
```

The `checkout` step is a bit confusing because the code fetches either
a branch or tag depending on if `CIRCLE_TAG` is defined.
Additionally, the code resets the `HEAD` multiple times. The only `git reset` that matters is that last one.  With some cleverness, we can

## A faster Git checkout with a shallow clone

The main speedup is to use a shallow clone with the `--depth=N`
flag. A shallow clone truncates the Git history to the specified
number of commits, typically 1 commit. CircleCI doesn’t offer a
shallow clone option on the builtin `checkout` step because GitHub
would prefer to [avoid the expensive computation](https://github.com/circleci/circleci-docs/issues/2040#issuecomment-368129275)
associated with shallow clones. I can’t speak to GitHub’s load but
it’s much faster for continuous integration to clone a shallow repo of
25MB than a full repo of 500MB.  The difference is 2 seconds instead
of 30 seconds.  For the full code, see the
`RUN_CHECKOUT_SHALLOW_GIT_REPO` alias in the example [CircleCI config](https://github.com/jschaf/ci_speed_test/blob/master/.circleci/config.yml). The
relevant section is below.

```bash
git init --quiet
git remote add origin "${CIRCLE_REPOSITORY_URL}"
# Fetch the repo contents at $CIRCLE_SHA1 directly into the local
# branch $BRANCH_NAME.
#
# --depth=1 for a shallow clone.
# --update-head-ok to allow updating the current HEAD. Occurs when
#     $CIRCLE_BRANCH is master.
# --force to always update the local branch.
git fetch --depth=1 --update-head-ok --force origin "${CIRCLE_SHA1}:${CIRCLE_BRANCH}"
git checkout --quiet ${CIRCLE_BRANCH}
```

The primary differences from the builtin CircleCI `checkout` step are:

* The `--depth=1` flag fetches a single commit.
* There’s only a single remote call to `git fetch` instead of two
  remote calls: `git clone; git fetch`.  The reason we can skip the
  second `git fetch` is because we’re fetching the repository at the
  precise Git hash we need to checkout from the
  `[CIRCLECI_SHA1](https://circleci.com/docs/2.0/env-vars/)`
  environmental variable.

# Common errors and solutions

## Git refusing to fetch into current branch of non-bare repository

Git displays the following error if you try to fetch the currently
checked out branch. For the specific error below, the current local
branch was `master` and the command run was `git fetch origin ${CIRCLE_SHA1}:master`.

```bash
fatal: Refusing to fetch into current branch refs/heads/master of non-bare repository
```

The error occurs because `git init` creates and checks out the
`master` branch by default.  The `CIRCLE_BRANCH` is also `master` when
either running tests on the repository or when executing a CircleCI
workflow from a [Git
tag](https://circleci.com/docs/2.0/workflows/#executing-workflows-for-a-git-tag).

To fix the error, use `[git fetch --update-head-ok](https://git-scm.com/docs/git-fetch#Documentation/git-fetch.txt---update-head-ok)`
which allows updating the head of current branch.  The documentation
warns not to use the flag “unless you are implementing your own
porcelain.” In this case, we are implementing our own porcelain to
automate the Git checkout on a continuous integration machine.

## Fixing SSH warnings for Permanently added RSA host key

[https://stackoverflow.com/questions/57652797/git-warns-with-warning-permanently-added-to-the-list-of-known-hosts-despite-a/57652847#57652847](https://stackoverflow.com/questions/57652797/git-warns-with-warning-permanently-added-to-the-list-of-known-hosts-despite-a/57652847#57652847)
Warning: Permanently added the RSA host key for IP address '140.82.114.3' to the list of known hosts.
