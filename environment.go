package ssh2docker

import "fmt"

type Environment map[string]string

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
