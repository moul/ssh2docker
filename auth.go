package ssh2docker

import (
	"fmt"

	"golang.org/x/crypto/ssh"
)

// ImageIsAllowed returns true if the target image is in the allowed list
func (s *Server) ImageIsAllowed(target string) bool {
	if s.AllowedImages == nil {
		return true
	}
	for _, image := range s.AllowedImages {
		if image == target {
			return true
		}
	}
	return false
}

// PasswordCallback is called when the user tries to authenticate using a password
func (s *Server) PasswordCallback(conn ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
	image := conn.User()
	if s.ImageIsAllowed(image) {
		return nil, nil
	}
	return nil, fmt.Errorf("TEST")
}
