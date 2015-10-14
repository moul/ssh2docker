package ssh2docker

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/moul/ssh2docker/vendor/github.com/Sirupsen/logrus"
	"github.com/moul/ssh2docker/vendor/golang.org/x/crypto/ssh"
)

// CheckConfig checks if the ClientConfig has access
func (s *Server) CheckConfig(config *ClientConfig) error {
	if !config.Allowed && s.PasswordAuthScript != "" {
		logrus.Warnf("config.Allowed = false")
		return fmt.Errorf("Access not allowed")
	}

	if s.AllowedImages != nil {
		allowed := false
		for _, image := range s.AllowedImages {
			if image == config.ImageName {
				allowed = true
				break
			}
		}
		if !allowed {
			logrus.Warnf("Image is not allowed: %q", config.ImageName)
			return fmt.Errorf("Image not allowed")
		}
	}

	return nil
}

// PublicKeyCallback is called when the user tries to authenticate using an SSH public key
func (s *Server) PublicKeyCallback(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
	username := conn.User()
	clientID := conn.RemoteAddr().String()
	keyText := strings.TrimSpace(string(ssh.MarshalAuthorizedKey(key)))
	logrus.Debugf("PublicKeyCallback: %q %q", username, keyText)
	// sessionID := conn.SessionID()

	config := s.ClientConfigs[clientID]
	if config == nil {
		s.ClientConfigs[clientID] = &ClientConfig{
			RemoteUser: username,
			ImageName:  username,
			Keys:       []string{},
			Env:        make(Environment, 0),
		}
	}
	config = s.ClientConfigs[clientID]
	config.Keys = append(config.Keys, keyText)
	return nil, s.CheckConfig(config)
}

// KeyboardInteractiveCallback is called after PublicKeyCallback
func (s *Server) KeyboardInteractiveCallback(conn ssh.ConnMetadata, challenge ssh.KeyboardInteractiveChallenge) (*ssh.Permissions, error) {
	username := conn.User()
	clientID := conn.RemoteAddr().String()
	logrus.Debugf("KeyboardInteractiveCallback: %q %q", username, challenge)

	config := s.ClientConfigs[clientID]
	if config == nil {
		s.ClientConfigs[clientID] = &ClientConfig{
			RemoteUser: username,
			ImageName:  username,
			Keys:       []string{},
			Env:        make(Environment, 0),
		}
	}
	config = s.ClientConfigs[clientID]

	if len(config.Keys) == 0 {
		logrus.Warnf("No user keys, continuing with password authentication")
		return nil, s.CheckConfig(config)
	}

	if s.PublicKeyAuthScript != "" {
		logrus.Debugf("%d keys received, trying to authenticate using hook script", len(config.Keys))
		script, err := expandUser(s.PublicKeyAuthScript)
		if err != nil {
			logrus.Warnf("Failed to expandUser: %v", err)
			return nil, err
		}
		args := append([]string{username}, config.Keys...)
		cmd := exec.Command(script, args...)
		// FIXME: redirect stderr to logrus
		cmd.Stderr = os.Stderr
		output, err := cmd.Output()
		if err != nil {
			logrus.Warnf("Failed to execute publickey-auth-script: %v", err)
			return nil, err
		}

		err = json.Unmarshal(output, &config)
		if err != nil {
			logrus.Warnf("Failed to unmarshal json %q: %v", string(output), err)
			return nil, err
		}
	} else {
		logrus.Debugf("%d keys received, but no hook script, continuing", len(config.Keys))
	}

	return nil, s.CheckConfig(config)
}

// PasswordCallback is called when the user tries to authenticate using a password
func (s *Server) PasswordCallback(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	username := conn.User()
	clientID := conn.RemoteAddr().String()

	logrus.Debugf("PasswordCallback: %q %q", username, password)

	config := s.ClientConfigs[clientID]
	if config == nil {
		s.ClientConfigs[clientID] = &ClientConfig{
			//Allowed: true,
			RemoteUser: username,
			ImageName:  username,
			Keys:       []string{},
			Env:        make(Environment, 0),
		}
	}
	config = s.ClientConfigs[clientID]

	if s.PasswordAuthScript != "" {
		script, err := expandUser(s.PasswordAuthScript)
		if err != nil {
			logrus.Warnf("Failed to expandUser: %v", err)
			return nil, err
		}
		cmd := exec.Command(script, username, string(password))
		// FIXME: redirect stderr to logrus
		cmd.Stderr = os.Stderr
		output, err := cmd.Output()
		if err != nil {
			logrus.Warnf("Failed to execute password-auth-script: %v", err)
			return nil, err
		}

		err = json.Unmarshal(output, &config)
		if err != nil {
			logrus.Warnf("Failed to unmarshal json %q: %v", string(output), err)
			return nil, err
		}
	}

	return nil, s.CheckConfig(config)
}
