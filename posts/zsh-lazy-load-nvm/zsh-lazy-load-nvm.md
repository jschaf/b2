+++
slug = "zsh-lazy-load"
date = 2019-03-29
visibility = "published"
+++

# Fix sluggish ZSH shells with lazy loading for slow scripts

I care immensely for having a snappy terminal startup and rebel against demands
to blindly source scripts in my ~/.zshrc.

The most recent challenge was Node Version Manager. NVM commits the mortal sin
of automatically adding itself to the ~/.zshrc file in the following
incantation:

```bash
URL='https://raw.githubusercontent.com/creationix/nvm/v0.34.0/install.sh'
curl -o- "${URL}" | bash

# Added to ~/.zshrc.
[ -s "$NVM_DIR/nvm.sh" ] && \. "$NVM_DIR/nvm.sh"
```

The NVM script requires 800ms to load, a debt paid every time I startup a shell.
Consulting internet canon, we calculate how much time it’s worth spending to
reclaim the 800ms of shell startup.

CONTINUE_READING

![Don't forget the time you spend finding the chart to look up what you save. 
And the time spent reading this reminder about the time spent. And the time 
trying to figure out if either of those actually make sense. Remember, every 
second counts toward your life total, including these right now.](./xkcd_1205_is_it_worth_the_time_orig.png)

> [xkcd-1205](https://xkcd.com/1205/): Is it worth the time? How long can you
> work on making a routine task more efficient before you're spending more time
> than you save? (across five years).

Assuming I open 20 shells a day, and it costs 800ms each time, we can
approximate from the table entry of 1 second, 50⁄day for a time of 1 day. Then,
scale down by 20⁄50 = 0.4 for shells per day and by 800⁄1000 = 0.8 for the 800ms
startup. xkcd allows us 0.32 days or about 8 hours to spend fixing this
transgression.

The solution is to lazy-load an nvm function that:

1. Unloads the fake nvm function.
2. Sources the slow NVM script to expose the real functionality.
3. Runs the real NVM.

```bash
# Shim NVM that will load the real NVM.
# Saves 800ms of script start-up time.
function nvm() {
  if [[ ! -e "${NVM_DIR}/nvm.sh" ]]; then
    echo "NVM not found in NVM_DIR=${NVM_DIR}" >&2
    return 1
  fi

  unfunction nvm
  source "${NVM_DIR}/nvm.sh"
  nvm "$@"
}
```
