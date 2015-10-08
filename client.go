package ssh2docker

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"

	"github.com/moul/ssh2docker/vendor/github.com/Sirupsen/logrus"
	"github.com/moul/ssh2docker/vendor/github.com/kr/pty"
	"github.com/moul/ssh2docker/vendor/golang.org/x/crypto/ssh"
)

var clientCounter = 0

// Client is one client connection
type Client struct {
	Idx        int
	ChannelIdx int
	Conn       *ssh.ServerConn
	Chans      <-chan ssh.NewChannel
	Reqs       <-chan *ssh.Request
	Server     *Server
	Pty, Tty   *os.File
	Env        Environment
	Config     *ClientConfig
	ClientID   string
}

type ClientConfig struct {
	ImageName  string `json:"image-name",omitempty`
	RemoteUser string `json:"remote-user",omitempty`
	Allowed    bool   `json:"allowed",omitempty`
}

// NewClient initializes a new client
func NewClient(conn *ssh.ServerConn, chans <-chan ssh.NewChannel, reqs <-chan *ssh.Request, server *Server) *Client {
	client := Client{
		Idx:        clientCounter,
		ClientID:   conn.RemoteAddr().String(),
		ChannelIdx: 0,
		Conn:       conn,
		Chans:      chans,
		Reqs:       reqs,
		Server:     server,
		Env: Environment{
			"TERM":              os.Getenv("TERM"),
			"DOCKER_HOST":       os.Getenv("DOCKER_HOST"),
			"DOCKER_CERT_PATH":  os.Getenv("DOCKER_CERT_PATH"),
			"DOCKER_TLS_VERIFY": os.Getenv("DOCKER_TLS_VERIFY"),
		},

		// Default ClientConfig, maybe we should completely remove it
		Config: &ClientConfig{
			ImageName:  conn.User(),
			RemoteUser: "anonymous",
		},
	}

	server.ClientConfigs[client.ClientID] = client.Config

	clientCounter++

	logrus.Infof("NewClient (%d): User=%q, ClientVersion=%q", client.Idx, conn.User(), fmt.Sprintf("%x", conn.ClientVersion()))
	return &client
}

// HandleRequests handles SSH requests
func (c *Client) HandleRequests() error {
	go func(in <-chan *ssh.Request) {
		for req := range in {
			logrus.Debugf("HandleRequest: %v", req)
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
		logrus.Debugf("Unknown channel type: %s", newChannel.ChannelType())
		newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
		return nil
	}

	channel, requests, err := newChannel.Accept()
	if err != nil {
		logrus.Errorf("newChannel.Accept failed: %v", err)
		return err
	}
	c.ChannelIdx++
	logrus.Debugf("HandleChannel.channel (client=%d channel=%d): %v", c.Idx, c.ChannelIdx, channel)

	logrus.Debug("Creating pty...")
	f, tty, err := pty.Open()
	if err != nil {
		logrus.Errorf("pty.Open failed: %v", err)
		return nil
	}
	c.Tty = tty
	c.Pty = f

	c.HandleChannelRequests(channel, requests)

	return nil
}

// HandleChannelRequests handles channel requests
func (c *Client) HandleChannelRequests(channel ssh.Channel, requests <-chan *ssh.Request) {
	go func(in <-chan *ssh.Request) {
		for req := range in {
			ok := false
			switch req.Type {
			case "shell":
				logrus.Debugf("HandleChannelRequests.req shell")
				if len(req.Payload) != 0 {
					break
				}
				ok = true

				// checking if a container already exists for this user
				existingContainer := ""
				if !c.Server.NoJoin {
					cmd := exec.Command("docker", "ps", "--filter=label=ssh2docker", fmt.Sprintf("--filter=label=image=%s", c.Config.ImageName), fmt.Sprintf("--filter=label=user=%s", c.Config.RemoteUser), "--quiet", "--no-trunc")
					buf, err := cmd.CombinedOutput()
					if err != nil {
						logrus.Warnf("docker ps ... failed: %v", err)
						continue
					}
					existingContainer = strings.TrimSpace(string(buf))
				}

				var cmd *exec.Cmd
				var err error

				// Opening Docker process
				if existingContainer != "" {
					// Attaching to an existing container
					args := []string{"exec", "-it", existingContainer, c.Server.DefaultShell}
					logrus.Debugf("Executing 'docker %s'", strings.Join(args, " "))
					cmd = exec.Command("docker", args...)
				} else {
					// Creating and attaching to a new container
					args := []string{"run"}
					args = append(args, c.Server.DockerRunArgs...)
					args = append(args, "--label=ssh2docker", fmt.Sprintf("--label=user=%s", c.Config.RemoteUser), fmt.Sprintf("--label=image=%s", c.Config.ImageName))
					args = append(args, c.Config.ImageName, c.Server.DefaultShell)
					logrus.Debugf("Executing 'docker %s'", strings.Join(args, " "))
					cmd = exec.Command("docker", args...)
					cmd.Env = c.Env.List()
				}

				defer c.Tty.Close()
				cmd.Stdout = c.Tty
				cmd.Stdin = c.Tty
				cmd.Stderr = c.Tty
				cmd.SysProcAttr = &syscall.SysProcAttr{
					Setctty: true,
					Setsid:  true,
				}

				err = cmd.Start()
				if err != nil {
					logrus.Warnf("cmd.Start failed: %v", err)
					continue
				}

				var once sync.Once
				close := func() {
					channel.Close()
					logrus.Debug("session closed")
				}

				go func() {
					io.Copy(channel, c.Pty)
					once.Do(close)
				}()

				go func() {
					io.Copy(c.Pty, channel)
					once.Do(close)
				}()

				go func() {
					if err := cmd.Wait(); err != nil {
						logrus.Warnf("cmd.Wait failed: %v", err)
					}
					once.Do(close)
				}()

			case "pty-req":
				ok = true
				termLen := req.Payload[3]
				c.Env["TERM"] = string(req.Payload[4 : termLen+4])
				w, h := parseDims(req.Payload[termLen+4:])
				SetWinsize(c.Pty.Fd(), w, h)
				logrus.Debugf("HandleChannelRequests.req pty-req: TERM=%q w=%q h=%q", c.Env["TERM"], int(w), int(h))

			case "window-change":
				w, h := parseDims(req.Payload)
				SetWinsize(c.Pty.Fd(), w, h)
				continue

			case "env":
				keyLen := req.Payload[3]
				key := string(req.Payload[4 : keyLen+4])
				valueLen := req.Payload[keyLen+7]
				value := string(req.Payload[keyLen+8 : keyLen+8+valueLen])
				logrus.Debugf("HandleChannelRequets.req 'env': %s=%q", key, value)
				c.Env[key] = value

			default:
				logrus.Debugf("Unhandled request type: %q: %v", req.Type, req)
			}

			if req.WantReply {
				if !ok {
					logrus.Debugf("Declining %s request...", req.Type)
				}
				req.Reply(ok, nil)
			}
		}
	}(requests)
}
