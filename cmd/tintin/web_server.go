package main

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/datatok/tintin/pkg/http"
)

func newWebServerCmd(out io.Writer) *cobra.Command {
	client := http.NewWebServer(settings)

	cmd := &cobra.Command{
		Use:     "server",
		Aliases: []string{"gen"},
		Short:   chartHelp,
		Long:    chartHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			return client.Run(out)
		},
	}

	f := cmd.Flags()

	f.IntVar(&client.Port, "port", 8080, "The port to listen on")

	return cmd
}
