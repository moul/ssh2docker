package ssh2docker

import "golang.org/x/crypto/ssh"

// PasswordCallback is called when the user tries to authenticate using a password
func (s *Server) PasswordCallback(conn ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
	return nil, nil
}
