# ssh2docker
:wrench: SSH server that can create new Docker containers and/or join existing ones, with session, and dynamic configuration support 

> SSH chroot with steroids

[![Build Status](https://travis-ci.org/moul/ssh2docker.svg?branch=master)](https://travis-ci.org/moul/ssh2docker)
[![GoDoc](https://godoc.org/github.com/moul/ssh2docker?status.svg)](https://godoc.org/github.com/moul/ssh2docker)
[![License](https://img.shields.io/github/license/moul/ssh2docker.svg)](https://github.com/moul/ssh2docker/blob/master/LICENSE)

![](https://raw.githubusercontent.com/moul/ssh2docker/master/resources/ssh2docker.png)

```ruby
┌────────────┐
│bobby@laptop│
└────────────┘
       │
       └──ssh container1@mycorp.biz──┐
                                     ▼
                               ┌──────────┐
┌──────────────────────────────┤ssh2docker├──┐
│                              └──────────┘  │
│              docker exec -it       │       │
│                 container1         │       │
│          ┌──────/bin/bash──────────┘       │
│ ┌────────┼───────────────────────────────┐ │
│ │docker  │                               │ │
│ │┌───────▼──┐ ┌──────────┐ ┌──────────┐  │ │
│ ││container1│ │container2│ │container3│  │ │
│ │└──────────┘ └──────────┘ └──────────┘  │ │
│ └────────────────────────────────────────┘ │
└────────────────────────────────────────────┘
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
   --verbose, -V                 Enable verbose mode
   --syslog-server               Configure a syslog server, i.e: udp://localhost:514
   --bind, -b ":2222"            Listen to address
   --host-key, -k "built-in"     Path or complete SSH host key to use, use 'system' for keys in /etc/ssh
   --allowed-images              List of allowed images, i.e: alpine,ubuntu:trusty,1cf3e6c
   --shell "/bin/sh"             DEFAULT shell
   --docker-run-args "-it --rm"  'docker run' arguments
   --no-join                     Do not join existing containers, always create new ones
   --clean-on-startup            Cleanup Docker containers created by ssh2docker on start
   --password-auth-script 	     Password auth hook file
   --publickey-auth-script 	     Public-key auth hook file
   --local-user 		         If setted, you can spawn a local shell (not withing docker) by SSHing to this user
   --banner 			         Display a banner on connection
   --help, -h			         show help
   --version, -v		         print the version
```

## Example

Server

```console
$ ssh2docker
INFO[0000] Listening on port 2222
INFO[0001] NewClient (0): User="alpine", ClientVersion="5353482d322e302d4f70656e5353485f362e362e317031205562756e74752d327562756e747532"
INFO[0748] NewClient (1): User="ubuntu", ClientVersion="5353482d322e302d4f70656e5353485f362e362e317031205562756e74752d327562756e747532"
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

## Test with Docker

You can test **ssh2docker** within Docker, but you will have some limitations, i.e: cannot run with boot2docker.

Here is an example about how to use ssh2docker inside Docker

```console
$ docker run --privileged -v /var/lib/docker:/var/lib/docker -it --rm -p 2222:2222 moul/ssh2docker
```

## Changelog

### master (unreleased)

* Support of `docker-exec-args` in hook scripts and in CLI args
* Sending environment variables to auth scripts
* TTY is now dynamic ([@quentinperez](https://github.com/quentinperez))
* Support of exec commands without tty, i.e: git-server, rsync, tftp, ...
* Support of API hooks for password and public key authentication ([#80](https://github.com/moul/ssh2docker/issues/80))
* Support of exec requests ([#51](https://github.com/moul/ssh2docker/issues/51))
* Support of `docker-run-args` in hook scripts ([#30](https://github.com/moul/ssh2docker/issues/30))
* Support of `--syslog-server` + refactored logs ([#71](https://github.com/moul/ssh2docker/issues/71))
* Do not ask for a password if only `--publickey-auth-script` is present ([#72](https://github.com/moul/ssh2docker/issues/72))
* Code refactor (split in modules), update examples, bump dependencies
* Support of `--syslog-server=unix:///dev/log` ([#74](https://github.com/moul/ssh2docker/issues/74))

[full commits list](https://github.com/moul/ssh2docker/compare/v1.2.0...master)

### [v1.2.0](https://github.com/moul/ssh2docker/releases/tag/v1.2.0) (2015-11-22)

* Support of `--host-key=system` to use OpenSSH keys ([#45](https://github.com/moul/ssh2docker/issues/45))
* Support of custom entrypoint ([#63](https://github.com/moul/ssh2docker/issues/63))
* Support of public-key authentication ([#2](https://github.com/moul/ssh2docker/issues/2))
* Handling custom environment variables, user and command in password script ([#57](https://github.com/moul/ssh2docker/issues/57))
* Replacing "_" by "/" on default image name to handle ControlMaster on clients
* Support of `--banner` option ([#26](https://github.com/moul/ssh2docker/issues/26))
* Add a not-yet-implemented warning for exec ([#51](https://github.com/moul/ssh2docker/issues/51))
* Support of `--local-user` option, to allow a specific user to be a local shell ([#44](https://github.com/moul/ssh2docker/issues/44))
* Kill connection when exiting shell (ctrl+D) ([#43](https://github.com/moul/ssh2docker/issues/43))

[full commits list](https://github.com/moul/ssh2docker/compare/v1.1.0...v1.2.0)

### [v1.1.0](https://github.com/moul/ssh2docker/releases/tag/v1.1.0) (2015-10-07)

* Fix runtime error on Linux ([#38](https://github.com/moul/ssh2docker/issues/38))
* Initial version of the native Scaleway support ([#36](https://github.com/moul/ssh2docker/issues/36))
* Support of 'ssh2docker --password-auth-script' options ([#28](https://github.com/moul/ssh2docker/issues/28))
* Add docker support ([#17](https://github.com/moul/ssh2docker/issues/17))
* Add GOXC support to build binaries for multiple architectures ([#18](https://github.com/moul/ssh2docker/issues/18))
* Support of 'ssh2docker --clean-on-startup' ([#23](https://github.com/moul/ssh2docker/issues/23))
* Add homebrew support ([#16](https://github.com/moul/ssh2docker/issues/16))
* Add Changelog ([#19](https://github.com/moul/ssh2docker/issues/19))

[full commits list](https://github.com/moul/ssh2docker/compare/v1.0.1...v1.1.0)

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
