package envhelper

import (
	"fmt"
	"os"
)

type Environment map[string]string

var defauldEnvVars = []string{"TERM", "DOCKER_HOST", "DOCKER_CERT_PATH", "DOCKER_TLS_VERIFY"}

func (e *Environment) List() []string {
	list := []string{}
	for k, v := range *e {
		if v == "" {
			continue
		}
		list = append(list, fmt.Sprintf("%s=%s", k, v))
	}
	return list
}

func (e *Environment) ApplyDefaults() {
	for _, name := range defauldEnvVars {
		if _, found := (*e)[name]; !found {
			(*e)[name] = os.Getenv(name)
		}
	}
}
