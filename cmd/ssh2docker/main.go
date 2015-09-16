package main

import (
	"net"
	"os"
	"path"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/moul/ssh2docker"
)

var VERSION string

func main() {
	app := cli.NewApp()
	app.Name = path.Base(os.Args[0])
	app.Author = "Manfred Touron"
	app.Email = "https://github.com/moul/ssh2docker"
	app.Version = VERSION
	app.Usage = "SSH portal to Docker containers"

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug, D",
			Usage: "Enable debug mode",
		},
		cli.StringFlag{
			Name:  "bind, b",
			Value: ":2222",
			Usage: "Listen to address",
		},
	}

	app.Action = Action

	app.Run(os.Args)
}

func hookBefore(c *cli.Context) error {
	// logrus.SetOutput(os.Stderr)
	if c.Bool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
	return nil
}

func Action(c *cli.Context) {
	server, err := ssh2docker.NewServer()
	if err != nil {
		logrus.Fatalf("Cannot create server: %v", err)
	}

	err = server.AddHostKeyFile("/Users/moul/Git/moul/ssh2docker/host_rsa")
	if err != nil {
		logrus.Fatalf("Cannot add host key file: %v", err)
	}

	bindAddress := c.String("bind")
	listener, err := net.Listen("tcp", bindAddress)
	if err != nil {
		logrus.Fatalf("Failed to start listener on %q: %v", bindAddress, err)
	}
	logrus.Infof("Listening on %q", bindAddress)

	for {
		conn, err := listener.Accept()
		if err != nil {
			logrus.Error("Accept failed: %v", err)
			continue
		}
		go server.Handle(conn)
	}
}
