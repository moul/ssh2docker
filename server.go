package ssh2docker

import (
	"net"

	"github.com/moul/ssh2docker/vendor/github.com/Sirupsen/logrus"
	"github.com/moul/ssh2docker/vendor/golang.org/x/crypto/ssh"
)

// Server is the ssh2docker main structure
type Server struct {
	SshConfig *ssh.ServerConfig
	// Clients   map[string]Client
	ClientConfigs map[string]*ClientConfig

	AllowedImages      []string
	DefaultShell       string
	DockerRunArgs      []string
	NoJoin             bool
	CleanOnStartup     bool
	PasswordAuthScript string
	LocalUser          string
	Banner             string

	initialized bool
}

// NewServer initialize a new Server instance with default values
func NewServer() (*Server, error) {
	server := Server{}
	server.SshConfig = &ssh.ServerConfig{
		PasswordCallback: server.PasswordCallback,
	}
	server.ClientConfigs = make(map[string]*ClientConfig, 0)
	server.DefaultShell = "/bin/sh"
	server.DockerRunArgs = []string{"-it", "--rm"}
	return &server, nil
}

// Init initializes server
func (s *Server) Init() error {
	// Initialize only once
	if s.initialized {
		return nil
	}

	if s.CleanOnStartup {
		err := DockerCleanup()
		if err != nil {
			logrus.Warnf("Failed to cleanup docker containers: %v", err)
		}
	}
	s.initialized = true
	return nil
}

// Handle is the SSH client entrypoint, it takes a net.Conn
// instance and handle all the ssh and ssh2docker stuff
func (s *Server) Handle(netConn net.Conn) error {
	if err := s.Init(); err != nil {
		return err
	}

	logrus.Debugf("Server.Handle netConn=%v", netConn)
	// Initialize a Client object
	conn, chans, reqs, err := ssh.NewServerConn(netConn, s.SshConfig)
	if err != nil {
		return err
	}
	client := NewClient(conn, chans, reqs, s)
	client.Config = s.ClientConfigs[conn.RemoteAddr().String()]

	// Handle requests
	if err = client.HandleRequests(); err != nil {
		return err
	}

	// Handle channels
	if err = client.HandleChannels(); err != nil {
		return err
	}
	return nil
}
