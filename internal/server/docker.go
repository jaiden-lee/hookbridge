package server

import (
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
)

func startTunnel(tunnelName string) (int, error) {
	nameEnvVar := "TUNNEL_NAME=" + tunnelName

	// temporary container for testing
	cmd := exec.Command(
		"docker", "run", "--rm", "-d",
		"--name", tunnelName,
		"-e", nameEnvVar,
		"--add-host=host.docker.internal:host-gateway",
		"-p", "0:50051", // docker chooses free port
		"tunnel-image",
	)

	out, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("docker run failed: %w\n%s", err, string(out))
	}

	// get port number
	cmd = exec.Command(
		"docker", "inspect", "-f",
		"{{(index (index .NetworkSettings.Ports \"50051/tcp\") 0).HostPort}}",
		tunnelName,
	)

	out, err = cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("docker port failed: %w\n%s", err, string(out))
	}

	portString := strings.TrimSpace(string(out))
	portNumber, err := strconv.Atoi(portString)
	if err != nil {
		return 0, fmt.Errorf("failed to parse port number: %w", err)
	}

	log.Println("Started container on port", portNumber)

	return portNumber, nil
}
