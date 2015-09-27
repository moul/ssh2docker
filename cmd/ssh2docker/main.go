package main

import (
	"net"
	"os"
	"path"
	"strings"

	"github.com/moul/ssh2docker"
	"github.com/moul/ssh2docker/vendor/github.com/Sirupsen/logrus"
	"github.com/moul/ssh2docker/vendor/github.com/codegangsta/cli"
)

var VERSION string

// Default key to ease getting started with the server
// You can easily use your own key by setting up
// `--host-key=/path/to/id_rsa`.
// See `man 1 ssh-keygen`.
const DefaultHostKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEA0BGfZSFn5ueRzMGPnd4+QkbrJ5vRmRXdg0D3ukSxFQC+QCXM
5lVYDtqp6DSsiIIr3PB0n94onebdDK7763RO1/fJP0ZBN3Ih1q0oQ9llrq7kuOd/
ttjNviu9KAkVQLcfR6zxttIPu/xnwkn7Y0pNOpn6ytjA2whemEKTAyskLSNVBqtW
r2TY/am7aXYG+1HSkbfSTSKI4ekzHAFLAZGK1q4FDOMAs6kC4IEmop1T3O2LvPBF
QzTt2WT0kph4+4saMqo4yoKEcKbdnWRkZul1YOcVyJReFX78fCKGo9tVjwtHHa3C
98WgjUiAXN2boGY2tPqk5vrTVmB69CJ6reezKQIDAQABAoIBAEBtRILnDio0iDPz
t4m1mGejWAtCt2sElzueMVcPEBolycNJMSIdSRAIa1YIgWgfjn9yQVqDSuZh5w6X
XFAzCnrbMgiSs3z8rTexFGe1+ENXymDq5ePzS/nXx1GPRnJsgZYLGil28AJQjLxf
diTvi+xaY4rOBSGNfOT+sFDp2eDTofTbxidgDzEJe0tWMi9QHs5NkyERYO7cpd6I
uXe0QLMP0aBzMKH4BFMiyBcY2gxYxqr1rC7YmEmn6M2HKxsHjGJ9K7msyJ0CdXNk
tiqwi3++T3jOcOz3u1t921/wUb4P7TJmerduEt5fm6wJVPCwZdQBcmoMznC0FGb2
5UwxeaUCgYEA7PTCQi7nxJB56CTxgM6m3D/+Lzdq1SBgUt3kzKPtJtDAfSCBd4hL
NYNyJ2WqJBdoclY+FQq85Rn2EdZY7Baol3WK26y+xWgXrICzPAtBvNByrr7eFpJS
2c35PrEuvJvsx0yn2HM4a8JYAm2yq+iJuW0aUNquhFLECvivsCfmXjcCgYEA4MqG
Rxy2Xu47UELfG3GONe6qd9gv2WWMzXtGXQfzYcYQMFUtfAj04/5ER92BEVqVTARB
ZOhdgLPTHsP4FwcjsCpjZQ2cW6QXUQfGJGrCWWd+3tFSv7ekrp4mXl4iXT/zJnWB
BrbABDS8qwz2+YhV6ql2wlYbVI7wdAf4tev2yZ8CgYBQNWmsTYRWnTEmy5qUJ1+E
HoVEJlYbXqI8arAQNU0JXpBJyr8IXzJWIvB5NYiqPuI0Ec1iAgh+5JLO5ueiwui+
nCMsyQSqfdnFoqsJICZYa5bmX+V9bnptD7PW7NMNNRqpO+F0+0uV7mssJ0XbuxMj
mTLXO67nS7zgmd2em2L3cQKBgAFYNMVoHo8izagFPmBjpX4dF1fwKxkZymXQPvN/
gK0tChu/5q2/P/e9JZtob8UyzYHO5LU9zpFegfzFH07D9CqxljachjrmGF2btkux
d8ghHlkm11/eMVX6DDC0T3BPWZz5RvRLU4qy5g3/3dpQPnNQ4Cz5ZuBymm2XPp2X
87nxAoGBAIh+WKnvopNektUbJX7hDE1HBVVaFfM3VffNDfBYF9ziPotRqY+th30N
Nn/aVy1yLb3M0eZCFD9A9/NE23kXAIFMAjM44NpqtXU/Z6Dmc+qB0Lk+KSL72cQL
hLP2sDRnSOAAGEYHcN75mvdrKahDRQXD8l0FIRUn/BNTlob8g5Tp
-----END RSA PRIVATE KEY-----`

func main() {
	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Author = "Manfred Touron"
	app.Email = "https://github.com/moul/ssh2docker"
	app.Version = VERSION
	app.Usage = "SSH portal to Docker containers"

	app.Before = hookBefore

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, V",
			Usage: "Enable verbose mode",
		},
		cli.StringFlag{
			Name:  "bind, b",
			Value: ":2222",
			Usage: "Listen to address",
		},
		cli.StringFlag{
			Name:  "host-key, k",
			Usage: "Path or complete SSH host key to use",
			Value: "built-in",
		},
		cli.StringFlag{
			Name:  "allowed-images",
			Usage: "List of allowed images, i.e: alpine,ubuntu:trusty,1cf3e6c",
			Value: "",
		},
		cli.StringFlag{
			Name:  "shell",
			Usage: "Default shell",
			Value: "/bin/sh",
		},
		cli.StringFlag{
			Name:  "docker-run-args",
			Usage: "'docker run' arguments",
			Value: "-it --rm",
		},
		cli.BoolFlag{
			Name:  "no-join",
			Usage: "Do not join existing containers, always create new ones",
		},
	}

	app.Action = Action

	app.Run(os.Args)
}

func hookBefore(c *cli.Context) error {
	// logrus.SetOutput(os.Stderr)
	if c.Bool("verbose") {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
	return nil
}

// Action is the default cli action to execute
func Action(c *cli.Context) {
	// Initialize the SSH server
	server, err := ssh2docker.NewServer()
	if err != nil {
		logrus.Fatalf("Cannot create server: %v", err)
	}

	// Restrict list of allowed images
	if c.String("allowed-images") != "" {
		server.AllowedImages = strings.Split(c.String("allowed-images"), ",")
	}

	// Configure server
	server.DefaultShell = c.String("shell")
	server.DockerRunArgs = strings.Split(c.String("docker-run-args"), " ")
	server.NoJoin = c.Bool("no-join")

	// Register the SSH host key
	hostKey := c.String("host-key")
	if hostKey == "built-in" {
		hostKey = DefaultHostKey
	}
	err = server.AddHostKey(hostKey)
	if err != nil {
		logrus.Fatalf("Cannot add host key: %v", err)
	}

	// Bind TCP socket
	bindAddress := c.String("bind")
	listener, err := net.Listen("tcp", bindAddress)
	if err != nil {
		logrus.Fatalf("Failed to start listener on %q: %v", bindAddress, err)
	}
	logrus.Infof("Listening on %q", bindAddress)

	// Accept new clients
	for {
		conn, err := listener.Accept()
		if err != nil {
			logrus.Error("Accept failed: %v", err)
			continue
		}
		go server.Handle(conn)
	}
}
