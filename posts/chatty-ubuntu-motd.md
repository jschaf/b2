+++
slug = "chatty-ubuntu-motd"
date = 2020-05-03
visibility = "published"
+++

# Cutting down the Ubuntu MOTD down to size

The Ubuntu message of the day (MOTD) is a chatty affair. A [MOTD][motd-wiki]
sends information to all users on login. A recent login message greeted me with
42 lines of questionable value.

[motd-wiki]: https://en.wikipedia.org/wiki/Motd_(Unix)

::: preview https://en.wikipedia.org/wiki/Motd_(Unix)
motd (Unix)

The **/etc/motd** is a file on Unix-like systems that contains a "message of the day", used
to send a common message to all users in a more efficient manner than sending them all an e-mail
message.
:::

The /etc/motd is a file on Unix-like systems that contains a "message of the day", used to send a
common message to all users in a more efficient manner than sending them all an e-mail message.
Other systems might also have a motd feature, such as the motd info segment on MULTICS.[1]

```text
Welcome to Ubuntu 18.04.2 LTS (GNU/Linux 4.15.0-1021-aws x86_64)

 * Documentation:  https://help.ubuntu.com
 * Management:     https://landscape.canonical.com
 * Support:        https://ubuntu.com/advantage

  System information as of Mon Apr 20 01:01:24 UTC 2020

  System load:  28.97              Processes:             10
  Usage of /:   28.9% of 48.41GB   Users logged in:       0
  Memory usage: 61%                IP address for enp4s0: 10.0.101.001
  Swap usage:   0%

 * Kubernetes 1.18 GA is now available! See https://microk8s.io
   for docs or install it with:

     sudo snap install microk8s --channel=1.18 --classic

 * Multipass 1.1 adds proxy support for developers behind enterprise
   firewalls. Rapid prototyping for cloud operations just got easier.

     https://multipass.run/

  Get cloud support with Ubuntu Advantage Cloud Guest:
    http://www.ubuntu.com/business/services/cloud

 * Canonical Livepatch is available for installation.
   - Reduce system reboots and improve kernel security. Activate at:
     https://ubuntu.com/livepatch

99 packages can be updated.
1 update is a security update.

The programs included with the Ubuntu system are free software;
the exact distribution terms for each program are described in the
individual files in /usr/share/doc/*/copyright.

Ubuntu comes with ABSOLUTELY NO WARRANTY, to the extent permitted by
applicable law.

*** System restart required ***
Last login: Sun Apr 19 03:41:04 2020 from 192.168.0.1
```

CONTINUE_READING

## Disable the verbose parts of the message

The commands listed below reduce the verbosity of the MOTD from 42 lines to 11
lines.

```shell script
# Disable cron job that updates the Ubuntu news.
systemctl disable motd-news.timer

# Disable Ubuntu news.
sed -i -e 's/ENABLED=1/ENABLED=0/' /etc/default/motd-news

# Disable MOTD scripts by removing the executable bit so run-parts won't run
# them.
chmod -x \
    /etc/update-motd.d/10-help-text \
    /etc/update-motd.d/50-motd-news \
    /etc/update-motd.d/51-cloudguest \
    /etc/update-motd.d/80-livepatch \
    /etc/update-motd.d/91-release-upgrade \
    /etc/update-motd.d/95-hwe-eol \
    /etc/update-motd.d/98-reboot-required

# The legal notice gets appended to the MOTD by PAM. The simplest way
# to drop it from the MOTD is to delete the file.
rm -f /etc/legal
```

The trimmed message looks like:

```text
Welcome to Ubuntu 18.04.2 LTS (GNU/Linux 4.15.0-1021-aws x86_64)

  System information as of Mon Apr 20 01:40:59 UTC 2020

  System load:  19.49              Processes:             10
  Usage of /:   28.9% of 48.41GB   Users logged in:       1
  Memory usage: 50%                IP address for enp4s0: 10.0.101.001
  Swap usage:   0%

152 packages can be updated.
1 update is a security update.
```

## History of the Ubuntu MOTD

I'm not the first to discover the chatty MOTD. The inclusion of ads in the MOTD
briefly outraged Hacker News and caused several bloggers to blog.

- [Disable motd news or (parts of) the dynamic motd on Ubuntu][raymii] - Deep
  dive into disabling run-parts, the PAM MOTD, and tweaking the MOTD config.
- [How To Disable Ads In Terminal Welcome Message In Ubuntu Server][technix] - A
  shallow dive into disabling the Ubuntu ads.
- [HN: Ubuntu displays advertising in /etc/motd ][hn ubuntu]
- [HN: BSD vs. Ubuntu motd(5)][hn bsd]

[raymii]:
  https://raymii.org/s/tutorials/Disable_dynamic_motd_and_motd_news_spam_on_Ubuntu_18.04.html
[hn ubuntu]: https://news.ycombinator.com/item?id=14662088
[hn bsd]: https://news.ycombinator.com/item?id=21893481
[technix]:
  https://www.ostechnix.com/how-to-disable-ads-in-terminal-welcome-message-in-ubuntu-server/

The best overview of the mechanics behind the Ubuntu MOTD is by David Kuhl in
the AskUbuntu question: _[How is /etc/motd updated?][how-motd]_ To summarize,
Ubuntu generates the MOTD by combining data from four sources.

1. Scripts run by `run-parts` in `/etc/update.motd/` like `00-header` or
   `10-help-text`.
1. Binaries invoked by `/etc/pam.d/login`.
1. SSH options from `/etc/ssh/sshd_config`, namely `PrintLastLog`.
1. The `/etc/legal` file.

[how-motd]: https://askubuntu.com/a/513900/544100

## Script contents from `/etc/update.motd/`

For reference, here's the output of the stock scripts in `/etc/update.motd/`.

- /etc/update.motd/00-header

  ```text
  Welcome to Ubuntu 18.04.2 LTS (GNU/Linux 4.15.0-1021-aws x86_64)
  ```

- /etc/update.motd/10-help-text

  ```text
  * Documentation:  https://help.ubuntu.com
  * Management:     https://landscape.canonical.com
  * Support:        https://ubuntu.com/advantage
  ```

- /etc/update.motd/50-landscape-sysinfo

  ```text
  System information as of Mon Apr 20 01:15:45 UTC 2020

  System load:  20.6               Processes:             1031
  Usage of /:   28.9% of 48.41GB   Users logged in:       1
  Memory usage: 43%                IP address for enp4s0: 10.0.0.1
  Swap usage:   0%
  ```

- /etc/update.motd/50-motd-news

  ```text
  * Kubernetes 1.18 GA is now available! See https://microk8s.io
    for docs or install it with:

      sudo snap install microk8s --channel=1.18 --classic

  * Multipass 1.1 adds proxy support for developers
    behind enterprise firewalls. Rapid prototyping for
    cloud operations just got easier.

      https://multipass.run/
  ```

- /etc/update.motd/51-cloudguest

  ```text
  Get cloud support with Ubuntu Advantage Cloud Guest:
    http://www.ubuntu.com/business/services/cloud
  ```

- /etc/update.motd/80-livepatch

  ```text
  Get cloud support with Ubuntu Advantage Cloud Guest:
    http://www.ubuntu.com/business/services/cloud
  ```

- /etc/update.motd/90-updates-available

  ```text
  152 packages can be updated.
  1 update is a security update.
  ```

### Legal text in the MOTD

The MOTD includes legal text in the login message. The simplest way to prevent
it from appearing in the MOTD is to delete the file: `rm /etc/legal`. The file
contains the following text:

```text
The programs included with the Ubuntu system are free software;
the exact distribution terms for each program are described in the
individual files in /usr/share/doc/*/copyright.

Ubuntu comes with ABSOLUTELY NO WARRANTY, to the extent permitted by
applicable law.
```
