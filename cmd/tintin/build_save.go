package main

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/datatok/tintin/pkg/action"
	"github.com/datatok/tintin/pkg/reporting/output"
)

const buildSaveHelp = `
Generate the report_io, from Djobi jobs / stages log, and save it.
`

func newReportBuildAsSaveCmd(client *action.ReportBuild, out io.Writer) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "save",
		Short: buildSaveHelp,
		Long:  buildSaveHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			report := client.Run()

			output.NewStore().SaveReport(report)

			return nil
		},
	}

	return cmd
}
