package main

import (
	"context"
	"fmt"
	"hookbridge/internal/client"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

func main() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	var (
		name      string
		port      int
		serverURL string
	)

	rootCmd := &cobra.Command{
		Use:   "hookbridge",
		Short: "Hookbridge CLI",
	}

	connectCmd := &cobra.Command{
		Use:   "connect",
		Short: "Connect to a tunnel and proxy requests to a local HTTP server",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
			defer stop()

			cfg := client.ConnectConfig{
				Name:      name,
				LocalPort: port,
				ServerURL: serverURL,
			}

			return client.RunConnect(ctx, cfg)
		},
	}

	connectCmd.Flags().StringVar(&name, "name", "", "Tunnel name")
	connectCmd.Flags().IntVar(&port, "port", 0, "Local HTTP server port")
	connectCmd.Flags().StringVar(&serverURL, "server", "http://localhost:8080", "Hookbridge server base URL")
	_ = connectCmd.MarkFlagRequired("name")
	_ = connectCmd.MarkFlagRequired("port")

	rootCmd.AddCommand(connectCmd)

	return rootCmd
}
