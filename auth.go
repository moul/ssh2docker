package ssh2docker

import (
	"encoding/json"
	"fmt"
	"os/exec"

	"github.com/moul/ssh2docker/vendor/github.com/Sirupsen/logrus"
	"github.com/moul/ssh2docker/vendor/golang.org/x/crypto/ssh"
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
func (s *Server) PasswordCallback(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	username := conn.User()
	clientID := conn.RemoteAddr().String()

	logrus.Debugf("PasswordCallback: %q %q", username, password)
	var image string

	if s.PasswordAuthScript != "" {
		// Using a hook script
		script, err := expandUser(s.PasswordAuthScript)
		if err != nil {
			logrus.Warnf("Failed to expandUser: %v", err)
			return nil, err
		}
		cmd := exec.Command(script, username, string(password))
		output, err := cmd.CombinedOutput()
		if err != nil {
			logrus.Warnf("Failed to execute password-auth-script: %v", err)
			return nil, err
		}

		var config ClientConfig
		err = json.Unmarshal(output, &config)
		if err != nil {
			logrus.Warnf("Failed to unmarshal json %q: %v", string(output), err)
			return nil, err
		}
		s.ClientConfigs[clientID] = &config
		if config.Allowed == false {
			logrus.Warnf("Hook returned allowed:false")
			return nil, fmt.Errorf("Access not allowed")
		}

		return nil, nil
	} else {
		// Default behavior
		image = username
	}

	if s.ImageIsAllowed(image) {
		return nil, nil
	}

	return nil, fmt.Errorf("TEST")
}
