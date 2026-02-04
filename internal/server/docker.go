package server

import (
	"fmt"
	"log"
	"os/exec"
)

func startTunnel(tunnelName string) error {
	nameEnvVar := "TUNNEL_NAME=" + tunnelName
	// temporary container for testing
	cmd := exec.Command(
		"docker", "run", "--rm", "-d",
		"--name", tunnelName,
		"-e", nameEnvVar,
		"--add-host=host.docker.internal:host-gateway",
		"nginx",
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker run failed: %w\n%s", err, string(out))
	}

	log.Println("Started container...")

	return nil
}
