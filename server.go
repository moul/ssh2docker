package ssh2docker

import (
	"io/ioutil"
	"net"
	"os"
	"strings"

	"github.com/apex/log"
	"github.com/moul/ssh2docker/pkg/dockerhelper"
	"golang.org/x/crypto/ssh"
)

// Server is the ssh2docker main structure
type Server struct {
	SshConfig *ssh.ServerConfig
	// Clients   map[string]Client
	ClientConfigs map[string]*ClientConfig

	AllowedImages        []string
	DefaultShell         string
	DockerRunArgsInline  string
	DockerExecArgsInline string
	PasswordAuthScript   string
	PublicKeyAuthScript  string
	LocalUser            string
	Banner               string
	NoJoin               bool
	CleanOnStartup       bool

	initialized bool
}

// NewServer initialize a new Server instance with default values
func NewServer() (*Server, error) {
	server := Server{}
	server.SshConfig = &ssh.ServerConfig{
		PasswordCallback:            server.PasswordCallback,
		PublicKeyCallback:           server.PublicKeyCallback,
		KeyboardInteractiveCallback: server.KeyboardInteractiveCallback,
	}
	server.ClientConfigs = make(map[string]*ClientConfig, 0)
	server.DefaultShell = "/bin/sh"
	return &server, nil
}

// Init initializes server
func (s *Server) Init() error {
	// Initialize only once
	if s.initialized {
		return nil
	}

	// disable password authentication
	if s.PasswordAuthScript == "" && s.PublicKeyAuthScript != "" {
		s.SshConfig.PasswordCallback = nil
	}

	// cleanup old containers
	if s.CleanOnStartup {
		err := dockerhelper.DockerCleanup()
		if err != nil {
			log.Warnf("Failed to cleanup docker containers: %v", err)
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

	log.Debugf("Server.Handle netConn=%v", netConn)
	// Initialize a Client object
	conn, chans, reqs, err := ssh.NewServerConn(netConn, s.SshConfig)

	if err != nil {
		log.Infof("Received disconnect from %s: 11: Bye Bye [preauth]", netConn.RemoteAddr().String())
		return err
	}
	client := NewClient(conn, chans, reqs, s)

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

// AddHostKey parses/loads an ssh key and registers it to the server
func (s *Server) AddHostKey(keystring string) error {
	// Check if keystring is a key path or a key string
	keypath := os.ExpandEnv(strings.Replace(keystring, "~", "$HOME", 2))
	_, err := os.Stat(keypath)
	var keybytes []byte
	if err == nil {
		keybytes, err = ioutil.ReadFile(keypath)
		if err != nil {
			return err
		}
	} else {
		keybytes = []byte(keystring)
	}

	// Parse SSH priate key
	hostkey, err := ssh.ParsePrivateKey(keybytes)
	if err != nil {
		return err
	}

	// Register key to the server
	s.SshConfig.AddHostKey(hostkey)
	return nil
}
