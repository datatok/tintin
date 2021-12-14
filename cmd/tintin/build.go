package main

import (
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/datatok/tintin/pkg/action"
)

const chartHelp = `
Generate the report, from Djobi jobs / stages log.
`

func newReportBuildCmd(out io.Writer) *cobra.Command {
	client := action.NewBuildReport(settings)

	cmd := &cobra.Command{
		Use:     "build",
		Aliases: []string{"gen"},
		Short:   chartHelp,
		Long:    chartHelp,
	}

	flags := cmd.PersistentFlags()

	cmd.AddCommand(
		newReportBuildAsTableCmd(client, out),
		newReportBuildAsTemplateCmd(client, out),
		newReportBuildAsSaveCmd(client, out),
		newReportBuildAsEmailCmd(client, out),
		sendMetricsCmd(client, out),
	)

	dateDefault := time.Now().AddDate(0, 0, -1).Format("02/01/2006")

	flags.StringVar(&client.Filter.Schedule, "schedule", dateDefault, "Schedule title")
	flags.StringVar(&client.Filter.Pipelines, "filter_pipeline", "", "Select only some pipelines")
	flags.StringSliceVarP(&client.Filter.Status, "filter_status", "", []string{}, "Select only some level (success/warning/error)")

	return cmd
}
