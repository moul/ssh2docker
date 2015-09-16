package ssh2docker

import (
	"net"

	"golang.org/x/crypto/ssh"
)

// Server is the ssh2docker main structure
type Server struct {
	SshConfig *ssh.ServerConfig
	// Clients   []Client
}

// NewServer initialize a new Server instance with default values
func NewServer() (*Server, error) {
	server := Server{}
	server.SshConfig = &ssh.ServerConfig{
		PasswordCallback: server.PasswordCallback,
	}
	return &server, nil
}

// Handle is the SSH client entrypoint, it takes a net.Conn
// instance and handle all the ssh and ssh2docker stuff
func (s *Server) Handle(netConn net.Conn) error {
	// Initialize a Client object
	conn, chans, reqs, err := ssh.NewServerConn(netConn, s.SshConfig)
	if err != nil {
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
