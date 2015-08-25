# ssh2docker
:wrench: SSH server that creates a Docker container per connection (chroot++)

[![GoDoc](https://godoc.org/github.com/moul/ssh2docker?status.svg)](https://godoc.org/github.com/moul/ssh2docker)

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

## License

MIT
