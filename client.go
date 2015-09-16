package ssh2docker

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/kr/pty"
	"golang.org/x/crypto/ssh"
)

// Client is one client connection
type Client struct {
	Conn   *ssh.ServerConn
	Chans  <-chan ssh.NewChannel
	Reqs   <-chan *ssh.Request
	Server *Server
}

// NewClient initializes a new client
func NewClient(conn *ssh.ServerConn, chans <-chan ssh.NewChannel, reqs <-chan *ssh.Request, server *Server) *Client {
	client := Client{
		Conn:   conn,
		Chans:  chans,
		Reqs:   reqs,
		Server: server,
	}
	logrus.Infof("NewClient: User=%q, ClientVersion=%1", conn.User(), fmt.Sprintf("%x", conn.ClientVersion()))
	return &client
}

// HandleRequests handles SSH requests
func (c *Client) HandleRequests() error {
	go func(in <-chan *ssh.Request) {
		for req := range in {
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}(c.Reqs)
	return nil
}

// HandleChannels handles SSH channels
func (c *Client) HandleChannels() error {
	for newChannel := range c.Chans {
		if err := c.HandleChannel(newChannel); err != nil {
			return err
		}
	}
	return nil
}

// HandleChannel handles one SSH channel
func (c *Client) HandleChannel(newChannel ssh.NewChannel) error {
	if newChannel.ChannelType() != "session" {
		newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
		return nil
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
		return nil
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

				args := []string{"run", "-it", "--rm", c.Conn.User(), "/bin/sh"}
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
	return nil
}
