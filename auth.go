package ssh2docker

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/apex/log"
	"github.com/mitchellh/go-homedir"
	"github.com/moul/ssh2docker/pkg/envhelper"
	"github.com/parnurzeal/gorequest"
	"golang.org/x/crypto/ssh"
)

// CheckConfig checks if the ClientConfig has access
func (s *Server) CheckConfig(config *ClientConfig) error {
	if !config.Allowed && (s.PasswordAuthScript != "" || s.PublicKeyAuthScript != "") {
		log.Debugf("config.Allowed = false")
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
			log.Warnf("Image is not allowed: %q", config.ImageName)
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
	log.Debugf("PublicKeyCallback: %q %q", username, keyText)
	// sessionID := conn.SessionID()

	config := s.ClientConfigs[clientID]
	if config == nil {
		s.ClientConfigs[clientID] = &ClientConfig{
			RemoteUser:             username,
			ImageName:              username,
			Keys:                   []string{},
			AuthenticationMethod:   "noauth",
			AuthenticationAttempts: 0,
			AuthenticationComment:  "",
			Env: make(envhelper.Environment, 0),
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
	log.Debugf("KeyboardInteractiveCallback: %q", username)

	config := s.ClientConfigs[clientID]
	if config == nil {
		s.ClientConfigs[clientID] = &ClientConfig{
			RemoteUser:             username,
			ImageName:              username,
			Keys:                   []string{},
			AuthenticationMethod:   "noauth",
			AuthenticationAttempts: 0,
			AuthenticationComment:  "",
			Env: make(envhelper.Environment, 0),
		}
	}
	config = s.ClientConfigs[clientID]

	if len(config.Keys) == 0 {
		log.Warnf("No user keys, continuing with password authentication")
		return nil, s.CheckConfig(config)
	}

	if s.PublicKeyAuthScript == "" {
		log.Debugf("%d keys received, but no hook script, continuing", len(config.Keys))
		return nil, s.CheckConfig(config)
	}

	config.AuthenticationAttempts++
	log.Debugf("%d keys received, trying to authenticate using publickey hook", len(config.Keys))

	var output []byte
	switch {
	case strings.HasPrefix(s.PublicKeyAuthScript, "http://"),
		strings.HasPrefix(s.PublicKeyAuthScript, "https://"):
		input := struct {
			Username   string   `json:"username"`
			Publickeys []string `json:"publickeys"`
		}{
			Username:   username,
			Publickeys: config.Keys,
		}
		resp, body, errs := gorequest.New().Type("json").Post(s.PublicKeyAuthScript).Send(input).End()
		if len(errs) > 0 {
			return nil, fmt.Errorf("gorequest errs: %v", errs)
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("invalid status code: %d", resp.StatusCode)
		}
		output = []byte(body)
	default:
		script, err := homedir.Expand(s.PublicKeyAuthScript)
		if err != nil {
			log.Warnf("Failed to expandUser: %v", err)
			return nil, err
		}
		cmd := exec.Command(script, append([]string{username}, config.Keys...)...)
		cmd.Env = config.Env.List()
		// FIXME: redirect stderr to log
		cmd.Stderr = os.Stderr
		output, err = cmd.Output()
		if err != nil {
			log.Warnf("Failed to execute publickey-auth-script: %v", err)
			return nil, err
		}
	}

	if err := json.Unmarshal(output, &config); err != nil {
		log.Warnf("Failed to unmarshal json %q: %v", string(output), err)
		return nil, err
	}

	if err := s.CheckConfig(config); err != nil {
		return nil, err
	}

	// success
	config.AuthenticationMethod = "publickey"
	return nil, nil
}

// PasswordCallback is called when the user tries to authenticate using a password
func (s *Server) PasswordCallback(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
	username := conn.User()
	clientID := conn.RemoteAddr().String()

	log.Debugf("PasswordCallback: %q %q", username, password)

	// map config in the memory
	config := s.ClientConfigs[clientID]
	if config == nil {
		s.ClientConfigs[clientID] = &ClientConfig{
			//Allowed: true,
			RemoteUser:             username,
			ImageName:              username,
			Keys:                   []string{},
			AuthenticationMethod:   "noauth",
			AuthenticationAttempts: 0,
			AuthenticationComment:  "",
			Env: make(envhelper.Environment, 0),
		}
		config = s.ClientConfigs[clientID]
	}

	// if there is a password callback
	if s.PasswordAuthScript == "" {
		return nil, s.CheckConfig(config)
	}

	config.AuthenticationAttempts++

	var output []byte
	switch {
	case strings.HasPrefix(s.PasswordAuthScript, "http://"),
		strings.HasPrefix(s.PasswordAuthScript, "https://"):
		input := struct {
			Username string `json:"username"`
			Password string `json:"password"`
		}{
			Username: username,
			Password: string(password),
		}
		resp, body, errs := gorequest.New().Type("json").Post(s.PasswordAuthScript).Send(input).End()
		if len(errs) > 0 {
			return nil, fmt.Errorf("gorequest errs: %v", errs)
		}
		if resp.StatusCode != 200 {
			return nil, fmt.Errorf("invalid status code: %d", resp.StatusCode)
		}
		output = []byte(body)
	default:
		script, err := homedir.Expand(s.PasswordAuthScript)
		if err != nil {
			log.Warnf("Failed to expandUser: %v", err)
			return nil, err
		}
		cmd := exec.Command(script, username, string(password))
		cmd.Env = config.Env.List()
		// FIXME: redirect stderr to log
		cmd.Stderr = os.Stderr
		output, err = cmd.Output()
		if err != nil {
			log.Warnf("Failed to execute password-auth-script: %v", err)
			return nil, err
		}
	}

	if err := json.Unmarshal(output, &config); err != nil {
		log.Warnf("Failed to unmarshal json %q: %v", string(output), err)
		return nil, err
	}

	if err := s.CheckConfig(config); err != nil {
		return nil, err
	}

	// success
	config.AuthenticationMethod = "password"
	return nil, nil
}
