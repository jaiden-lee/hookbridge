package server

import (
	"log"
	"os/exec"
)

func startTunnel() error {
	// temporary container for testing
	cmd := exec.Command(
		"docker", "run",
		"-it", "ubuntu", "bash",
	)

	err := cmd.Run()
	if err != nil {
		return err
	}

	log.Println("Started container...")
}
