package ssh2docker

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/kr/pty"
	"golang.org/x/crypto/ssh"
)

// Server is the ssh2docker main structure
type Server struct {
	sshConfig *ssh.ServerConfig
}

// NewServer initialize a new Server instance with default values
func NewServer() (*Server, error) {
	server := Server{}
	server.sshConfig = &ssh.ServerConfig{
		PasswordCallback: server.PasswordCallback,
	}
	return &server, nil
}

// Handle is the SSH client entrypoint, it takes a net.Conn
// instance and handle all the ssh and ssh2docker stuff
func (s *Server) Handle(netConn net.Conn) error {
	conn, chans, reqs, err := ssh.NewServerConn(netConn, s.sshConfig)
	if err != nil {
		logrus.Errorf("%s", err)
		return err
	}

	logrus.Infof("conn: User=%q, ClientVersion=%1", conn.User(), fmt.Sprintf("%x", conn.ClientVersion()))

	go func(in <-chan *ssh.Request) {
		for req := range in {
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			logrus.Errorf("%s", err)
			return err
		}

		logrus.Infof("Creating pty...")
		f, tty, err := pty.Open()
		if err != nil {
			logrus.Errorf("%s", err)
			continue
		}

		termEnv := "xterm"

		go func(in <-chan *ssh.Request) {
			for req := range in {
				ok := false
				switch req.Type {
				case "shell":
					if len(req.Payload) != 0 {
						break
					}
					ok = true

					args := []string{"run", "-it", "--rm", conn.User(), "/bin/sh"}
					logrus.Infof("Executing docker %s", args)
					cmd := exec.Command("docker", args...)
					cmd.Env = []string{
						"TERM=" + termEnv,
						"DOCKER_HOST=" + os.Getenv("DOCKER_HOST"),
						"DOCKER_CERT_PATH=" + os.Getenv("DOCKER_CERT_PATH"),
						"DOCKER_TLS_VERIFY=" + os.Getenv("DOCKER_TLS_VERIFY"),
					}

					defer tty.Close()
					cmd.Stdout = tty
					cmd.Stdin = tty
					cmd.Stderr = tty
					cmd.SysProcAttr = &syscall.SysProcAttr{
						Setctty: true,
						Setsid:  true,
					}

					err := cmd.Start()
					if err != nil {
						logrus.Warnf("%s", err)
						continue
					}

					var once sync.Once
					close := func() {
						channel.Close()
						logrus.Infof("session closed")
					}

					go func() {
						io.Copy(channel, f)
						once.Do(close)
					}()

					go func() {
						io.Copy(f, channel)
						once.Do(close)
					}()

				case "pty-req":
					ok = true
					termLen := req.Payload[3]
					termEnv = string(req.Payload[4 : termLen+4])
					w, h := parseDims(req.Payload[termLen+4:])
					SetWinsize(f.Fd(), w, h)
					logrus.Infof("pty-req: %s", termEnv)

				case "window-changed":
					w, h := parseDims(req.Payload)
					SetWinsize(f.Fd(), w, h)
					continue
				}

				if req.WantReply {
					if !ok {
						logrus.Infof("Declining %s request...", req.Type)
					}
					req.Reply(ok, nil)
				}
			}
		}(requests)
	}

	return nil
}
