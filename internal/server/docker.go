package server

import (
	"fmt"
	"log"
	"os/exec"
)

func startTunnel(tunnelName string, port int) error {
	nameEnvVar := "TUNNEL_NAME=" + tunnelName
	portMapping := fmt.Sprintf("%d:50051", port)
	// temporary container for testing
	cmd := exec.Command(
		"docker", "run", "-d",
		"--name", tunnelName,
		"-e", nameEnvVar,
		"--add-host=host.docker.internal:host-gateway",
		"-p", portMapping,
		"tunnel-image",
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("docker run failed: %w\n%s", err, string(out))
	}

	log.Println("Started container...")

	return nil
}
