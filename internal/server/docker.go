package server

import (
	"log"
	"os/exec"
)

func startTunnel(tunnelName string) error {
	// temporary container for testing
	cmd := exec.Command(
		"docker", "run",
		"-it", "--name", tunnelName,
		"ubuntu", "bash",
	)

	err := cmd.Run()
	if err != nil {
		return err
	}

	log.Println("Started container...")
}
