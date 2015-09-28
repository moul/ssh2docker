# ssh2docker
:wrench: SSH server that creates a Docker container per connection (chroot++)

[![Build Status](https://travis-ci.org/moul/ssh2docker.svg?branch=master)](https://travis-ci.org/moul/ssh2docker)
[![GoDoc](https://godoc.org/github.com/moul/ssh2docker?status.svg)](https://godoc.org/github.com/moul/ssh2docker)
[![License](https://img.shields.io/github/license/moul/ssh2docker.svg)](https://github.com/moul/ssh2docker/blob/master/LICENSE)

![](https://raw.githubusercontent.com/moul/ssh2docker/master/resources/ssh2docker.png)

## Example

Server

```console
$ ssh2docker
INFO[0000] Listening on port 2222
INFO[0004] conn: User="alpine", ClientVersion=%!(NOVERB)%!(EXTRA string=5353482d322e302d4f70656e5353485f362e32)
INFO[0004] Creating pty...
INFO[0004] Window resize 181x50
INFO[0004] pty-req: xterm-256color
INFO[0004] Executing docker [run -it --rm alpine /bin/sh]
INFO[0010] session closed
INFO[0016] conn: User="ubuntu", ClientVersion=%!(NOVERB)%!(EXTRA string=5353482d322e302d4f70656e5353485f362e32)
INFO[0016] Creating pty...
INFO[0016] Window resize 181x50
INFO[0016] pty-req: xterm-256color
INFO[0016] Executing docker [run -it --rm ubuntu /bin/sh]
INFO[0023] session closed
```

Client

```console
$ ssh localhost -p 2222 -l alpine
Host key fingerprint is 59:46:d7:cf:ca:33:be:1f:58:fd:46:c8:ca:5d:56:03
+--[ RSA 2048]----+
|          . .E   |
|         . .  o  |
|          o    +.|
|         +   . .*|
|        S    .oo=|
|           . oB+.|
|            oo.+o|
|              ...|
|              .o.|
+-----------------+

alpine@localhost's password:
/ # cat /etc/alpine-release
3.2.0
/ # ^D
```

```console
$ ssh localhost -p 2222 -l ubuntu
Host key fingerprint is 59:46:d7:cf:ca:33:be:1f:58:fd:46:c8:ca:5d:56:03
+--[ RSA 2048]----+
|          . .E   |
|         . .  o  |
|          o    +.|
|         +   . .*|
|        S    .oo=|
|           . oB+.|
|            oo.+o|
|              ...|
|              .o.|
+-----------------+

ubuntu@localhost's password:
# lsb_release -a
No LSB modules are available.
Distributor ID:	Ubuntu
Description:	Ubuntu 14.04.3 LTS
Release:	14.04
Codename:	trusty
# ^D
```

## Install

Install latest version using Golang (recommended)

```console
$ go get github.com/moul/ssh2docker/cmd/ssh2docker
```

---

Install latest version using Homebrew (Mac OS X)

```console
$ brew install https://raw.githubusercontent.com/moul/ssh2docker/master/contrib/homebrew/assh.rb --HEAD

```

or the latest released version

```console
$ brew install https://raw.githubusercontent.com/moul/ssh2docker/master/contrib/homebrew/assh.rb

```

## Usage

```
NAME:
   ssh2docker - SSH portal to Docker containers

USAGE:
   ssh2docker [global options] command [command options] [arguments...]

AUTHOR(S):
   Manfred Touron <https://github.com/moul/ssh2docker>

COMMANDS:
   help, h	Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --verbose, -V		Enable verbose mode
   --bind, -b ":2222"		Listen to address
   --host-key, -k "built-in"	Path or complete SSH host key to use
   --allowed-images 		List of allowed images, i.e: alpine,ubuntu:trusty,1cf3e6c
   --shell "/bin/sh"		Default shell
   --docker-run-args "-it --rm"	'docker run' arguments
   --help, -h			show help
   --version, -v		print the version
```

## Changelog

### master (unreleased)

* Support of 'ssh2docker --clean-on-startup' ([#23](https://github.com/moul/ssh2docker/issues/23))
* Add homebrew support ([#16](https://github.com/moul/ssh2docker/issues/16))
* Add Changelog ([#19](https://github.com/moul/ssh2docker/issues/19))

[full commits list](https://github.com/moul/ssh2docker/compare/v1.0.1...master)

### [v1.0.1](https://github.com/moul/ssh2docker/releases/tag/v1.0.1) (2015-09-27)

* Using [party](https://github.com/mjibson/party) to manage dependencies

[full commits list](https://github.com/moul/ssh2docker/compare/v1.0.0...v1.0.1)

### [v1.0.0](https://github.com/moul/ssh2docker/releases/tag/v1.0.0) (2015-09-27)

**Initial release**

#### Features

* Basic logging
* Handling environment-variable requests
* Support of `--allowed-images` option ([#4](https://github.com/moul/ssh2docker/issues/4))
* Ability to configure `docker run` arguments ([#13](https://github.com/moul/ssh2docker/issues/13))
* Reconnecting to existing containers ([#14](https://github.com/moul/ssh2docker/issues/14))
* Support of `--no-join` option ([#6](https://github.com/moul/ssh2docker/issues/6))

[full commits list](https://github.com/moul/ssh2docker/compare/a398db225cefe1d1de642217be1c06d6c5d721b0...v1.0.0)

## License

MIT
