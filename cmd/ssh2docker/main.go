package main

import (
	"net"

	"github.com/Sirupsen/logrus"
	"github.com/moul/ssh2docker"
)

func main() {
	server, err := ssh2docker.NewServer()
	if err != nil {
		logrus.Fatalf("Cannot create server: %v", err)
	}

	err = server.AddHostKeyFile("/Users/moul/Git/moul/ssh2docker/host_rsa")
	if err != nil {
		logrus.Fatalf("Cannot add host key file: %v", err)
	}

	listener, err := net.Listen("tcp", ":2222")
	if err != nil {
		logrus.Fatalf("Failed to start listener: %v", err)
	}
	logrus.Infof("Listening on port 2222")

	for {
		conn, err := listener.Accept()
		if err != nil {
			logrus.Error("Accept failed: %v", err)
			continue
		}
		go server.Handle(conn)
	}
}
