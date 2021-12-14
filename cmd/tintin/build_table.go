package main

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/datatok/tintin/pkg/action"
	"github.com/datatok/tintin/pkg/reporting/output"
)

const buildTableHelp = `
Generate the report_io, from Djobi jobs / stages log.
`

func newReportBuildAsTableCmd(client *action.ReportBuild, out io.Writer) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "table",
		Short: buildTableHelp,
		Long:  buildTableHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			report := client.Run()

			output.ToTable(out, report)

			return nil
		},
	}

	return cmd
}
