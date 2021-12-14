package main

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/datatok/tintin/pkg/action"
	"github.com/datatok/tintin/pkg/metrics"
	"github.com/datatok/tintin/pkg/reporting/sender"
)

const buildTemplateMetricsHelp = `
Send metrics to push-gateway.
`

func sendMetricsCmd(client *action.ReportBuild, out io.Writer) *cobra.Command {
	e := sender.Email{}

	cmd := &cobra.Command{
		Use:   "metrics",
		Short: buildTemplateMetricsHelp,
		Long:  buildTemplateMetricsHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			metrics.New(settings).Push(e.To)

			return nil
		},
	}

	f := cmd.Flags()

	f.StringVar(&e.To, "to", "", "Push gateway URL")

	return cmd
}
