package ssh2docker

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"text/template"

	"github.com/apex/log"
	"github.com/flynn/go-shlex"
	"github.com/kr/pty"
	"github.com/moul/ssh2docker/pkg/envhelper"
	"github.com/moul/ssh2docker/pkg/ttyhelper"
	"golang.org/x/crypto/ssh"
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
	Config     *ClientConfig
	ClientID   string
}

type ClientConfig struct {
	ImageName              string                `json:"image-name,omitempty"`
	RemoteUser             string                `json:"remote-user,omitempty"`
	Env                    envhelper.Environment `json:"env,omitempty"`
	Command                []string              `json:"command,omitempty"`
	DockerRunArgs          []string              `json:"docker-run-args,omitempty"`
	DockerExecArgs         []string              `json:"docker-exec-args,omitempty"`
	User                   string                `json:"user,omitempty"`
	Keys                   []string              `json:"keys,omitempty"`
	AuthenticationMethod   string                `json:"authentication-method,omitempty"`
	AuthenticationComment  string                `json:"authentication-coment,omitempty"`
	EntryPoint             string                `json:"entrypoint,omitempty"`
	AuthenticationAttempts int                   `json:"authentication-attempts,omitempty"`
	Allowed                bool                  `json:"allowed,omitempty"`
	IsLocal                bool                  `json:"is-local,omitempty"`
	UseTTY                 bool                  `json:"use-tty,omitempty"`
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

		// Default ClientConfig, will be overwritten if a hook is used
		Config: &ClientConfig{
			ImageName:              strings.Replace(conn.User(), "_", "/", -1),
			RemoteUser:             "anonymous",
			AuthenticationMethod:   "noauth",
			AuthenticationComment:  "",
			AuthenticationAttempts: 0,
			Env:     envhelper.Environment{},
			Command: make([]string, 0),
		},
	}

	if server.LocalUser != "" {
		client.Config.IsLocal = client.Config.ImageName == server.LocalUser
	}

	if _, found := server.ClientConfigs[client.ClientID]; !found {
		server.ClientConfigs[client.ClientID] = client.Config
	}

	client.Config = server.ClientConfigs[conn.RemoteAddr().String()]
	client.Config.Env.ApplyDefaults()

	clientCounter++

	remoteAddr := strings.Split(client.ClientID, ":")
	log.Infof("Accepted %s for %s from %s port %s ssh2: %s", client.Config.AuthenticationMethod, conn.User(), remoteAddr[0], remoteAddr[1], client.Config.AuthenticationComment)
	return &client
}

// HandleRequests handles SSH requests
func (c *Client) HandleRequests() error {
	go func(in <-chan *ssh.Request) {
		for req := range in {
			log.Debugf("HandleRequest: %v", req)
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
		log.Debugf("Unknown channel type: %s", newChannel.ChannelType())
		newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
		return nil
	}

	channel, requests, err := newChannel.Accept()
	if err != nil {
		log.Errorf("newChannel.Accept failed: %v", err)
		return err
	}
	c.ChannelIdx++
	log.Debugf("HandleChannel.channel (client=%d channel=%d)", c.Idx, c.ChannelIdx)

	log.Debug("Creating pty...")
	c.Pty, c.Tty, err = pty.Open()
	if err != nil {
		log.Errorf("pty.Open failed: %v", err)
		return nil
	}

	c.HandleChannelRequests(channel, requests)

	return nil
}

func (c *Client) alterArg(arg string) (string, error) {
	tmpl, err := template.New("run-args").Parse(arg)
	if err != nil {
		return "", err
	}

	var buff bytes.Buffer
	if err := tmpl.Execute(&buff, c.Config); err != nil {
		return "", err
	}

	return buff.String(), nil
}

func (c *Client) alterArgs(args []string) error {
	for idx, arg := range args {
		newArg, err := c.alterArg(arg)
		if err != nil {
			return err
		}
		args[idx] = newArg
	}
	return nil
}

func (c *Client) runCommand(channel ssh.Channel, entrypoint string, command []string) {
	var cmd *exec.Cmd
	var err error

	if c.Config.IsLocal {
		cmd = exec.Command(entrypoint, command...)
	} else {
		// checking if a container already exists for this user
		existingContainer := ""
		if !c.Server.NoJoin {
			cmd = exec.Command("docker", "ps", "--filter=label=ssh2docker", fmt.Sprintf("--filter=label=image=%s", c.Config.ImageName), fmt.Sprintf("--filter=label=user=%s", c.Config.RemoteUser), "--quiet", "--no-trunc")
			cmd.Env = c.Config.Env.List()
			buf, err := cmd.CombinedOutput()
			if err != nil {
				log.Warnf("docker ps ... failed: %v", err)
				channel.Close()
				return
			}
			existingContainer = strings.TrimSpace(string(buf))
		}

		// Opening Docker process
		if existingContainer != "" {
			// Attaching to an existing container
			args := []string{"exec"}
			if len(c.Config.DockerExecArgs) > 0 {
				args = append(args, c.Config.DockerExecArgs...)
				if err := c.alterArgs(args); err != nil {
					log.Errorf("Failed to execute template on args: %v", err)
					return
				}
			} else {
				inlineExec, err := c.alterArg(c.Server.DockerExecArgsInline)
				if err != nil {
					log.Errorf("Failed to execute template on arg: %v", err)
					return
				}
				execArgs, err := shlex.Split(inlineExec)
				if err != nil {
					log.Errorf("Failed to split arg %q: %v", inlineExec, err)
					return
				}
				args = append(args, execArgs...)
			}

			args = append(args, existingContainer)
			if entrypoint != "" {
				args = append(args, entrypoint)
			}
			args = append(args, command...)
			log.Debugf("Executing 'docker %s'", strings.Join(args, " "))
			cmd = exec.Command("docker", args...)
			cmd.Env = c.Config.Env.List()
		} else {
			// Creating and attaching to a new container
			args := []string{"run"}
			if len(c.Config.DockerRunArgs) > 0 {
				args = append(args, c.Config.DockerRunArgs...)
				if err := c.alterArgs(args); err != nil {
					log.Errorf("Failed to execute template on args: %v", err)
					return
				}
			} else {
				inlineRun, err := c.alterArg(c.Server.DockerRunArgsInline)
				if err != nil {
					log.Errorf("Failed to execute template on arg: %v", err)
					return
				}
				runArgs, err := shlex.Split(inlineRun)
				if err != nil {
					log.Errorf("Failed to split arg %q: %v", inlineRun, err)
					return
				}
				args = append(args, runArgs...)
			}

			args = append(args, "--label=ssh2docker", fmt.Sprintf("--label=user=%s", c.Config.RemoteUser), fmt.Sprintf("--label=image=%s", c.Config.ImageName))
			if c.Config.User != "" {
				args = append(args, "-u", c.Config.User)
			}
			if entrypoint != "" {
				args = append(args, "--entrypoint", entrypoint)
			}

			args = append(args, c.Config.ImageName)
			args = append(args, command...)
			log.Debugf("Executing 'docker %s'", strings.Join(args, " "))
			cmd = exec.Command("docker", args...)
			cmd.Env = c.Config.Env.List()
		}
	}

	if c.Server.Banner != "" {
		banner := c.Server.Banner
		banner = strings.Replace(banner, "\r", "", -1)
		banner = strings.Replace(banner, "\n", "\n\r", -1)
		fmt.Fprintf(channel, "%s\n\r", banner)
	}

	cmd.Stdout = channel
	cmd.Stdin = channel
	cmd.Stderr = channel
	var wg sync.WaitGroup
	if c.Config.UseTTY {
		cmd.Stdout = c.Tty
		cmd.Stdin = c.Tty
		cmd.Stderr = c.Tty

		wg.Add(1)
		go func() {
			io.Copy(channel, c.Pty)
			wg.Done()
		}()
		wg.Add(1)
		go func() {
			io.Copy(c.Pty, channel)
			wg.Done()
		}()
		defer wg.Wait()
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setctty: c.Config.UseTTY,
		Setsid:  true,
	}

	err = cmd.Start()
	if err != nil {
		log.Warnf("cmd.Start failed: %v", err)
		channel.Close()
		return
	}

	if err := cmd.Wait(); err != nil {
		log.Warnf("cmd.Wait failed: %v", err)
	}
	channel.Close()
	log.Debugf("cmd.Wait done")
}

// HandleChannelRequests handles channel requests
func (c *Client) HandleChannelRequests(channel ssh.Channel, requests <-chan *ssh.Request) {
	go func(in <-chan *ssh.Request) {
		defer c.Tty.Close()

		for req := range in {
			ok := false
			switch req.Type {
			case "shell":
				log.Debugf("HandleChannelRequests.req shell")
				if len(req.Payload) != 0 {
					break
				}
				ok = true

				entrypoint := ""
				if c.Config.EntryPoint != "" {
					entrypoint = c.Config.EntryPoint
				}

				var args []string
				if c.Config.Command != nil {
					args = c.Config.Command
				}

				if entrypoint == "" && len(args) == 0 {
					args = []string{c.Server.DefaultShell}
				}

				c.runCommand(channel, entrypoint, args)

			case "exec":
				command := string(req.Payload[4:])
				log.Debugf("HandleChannelRequests.req exec: %q", command)
				ok = true

				args, err := shlex.Split(command)
				if err != nil {
					log.Errorf("Failed to parse command %q: %v", command, args)
				}
				c.runCommand(channel, c.Config.EntryPoint, args)

			case "pty-req":
				ok = true
				c.Config.UseTTY = true
				termLen := req.Payload[3]
				c.Config.Env["TERM"] = string(req.Payload[4 : termLen+4])
				c.Config.Env["USE_TTY"] = "1"
				w, h := ttyhelper.ParseDims(req.Payload[termLen+4:])
				ttyhelper.SetWinsize(c.Pty.Fd(), w, h)
				log.Debugf("HandleChannelRequests.req pty-req: TERM=%q w=%q h=%q", c.Config.Env["TERM"], int(w), int(h))

			case "window-change":
				w, h := ttyhelper.ParseDims(req.Payload)
				ttyhelper.SetWinsize(c.Pty.Fd(), w, h)
				continue

			case "env":
				keyLen := req.Payload[3]
				key := string(req.Payload[4 : keyLen+4])
				valueLen := req.Payload[keyLen+7]
				value := string(req.Payload[keyLen+8 : keyLen+8+valueLen])
				log.Debugf("HandleChannelRequets.req 'env': %s=%q", key, value)
				c.Config.Env[key] = value

			default:
				log.Debugf("Unhandled request type: %q: %v", req.Type, req)
			}

			if req.WantReply {
				if !ok {
					log.Debugf("Declining %s request...", req.Type)
				}
				req.Reply(ok, nil)
			}
		}
	}(requests)
}
