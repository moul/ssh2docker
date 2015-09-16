package ssh2docker

import (
	"io/ioutil"
	"os"
	"strings"

	"golang.org/x/crypto/ssh"
)

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
	s.sshConfig.AddHostKey(hostkey)
	return nil
}
