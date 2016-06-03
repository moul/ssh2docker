package dockerhelper

import (
	"os/exec"
	"strings"

	"github.com/apex/log"
)

// DockerCleanup cleans all containers created by ssh2docker
func DockerCleanup() error {
	containers, err := DockerListContainers(false)
	if err != nil {
		return err
	}

	for _, cid := range containers {
		if err = DockerKill(cid); err != nil {
			log.Warnf("Failed to kill container %q: %v", cid, err)
		}
	}

	containers, err = DockerListContainers(true)
	if err != nil {
		return err
	}

	for _, cid := range containers {
		if err = DockerRemove(cid); err != nil {
			log.Warnf("Failed to remove container %q: %v", cid, err)
		}
	}

	return nil
}

// DockerKill kills a container
func DockerKill(containerID string) error {
	cmd := exec.Command("docker", "kill", "-s", "9", containerID)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	log.Debugf("Killed container: %q", containerID)
	return nil
}

// DockerRemove removes a container
func DockerRemove(containerID string) error {
	cmd := exec.Command("docker", "rm", "-f", containerID)
	_, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}
	log.Debugf("Deleted container: %q", containerID)
	return nil
}

// DockerListContainers lists containers created by ssh2docker
func DockerListContainers(all bool) ([]string, error) {
	command := []string{"docker", "ps", "--filter=label=ssh2docker", "--quiet", "--no-trunc"}
	if all {
		command = append(command, "-a")
	}
	cmd := exec.Command(command[0], command[1:]...)
	buf, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	containers := strings.Split(strings.TrimSpace(string(buf)), "\n")
	if containers[0] == "" {
		return nil, nil
	}
	return containers, nil
}
