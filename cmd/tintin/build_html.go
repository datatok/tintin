package main

import (
	"io"

	"github.com/spf13/cobra"

	"github.com/datatok/tintin/pkg/action"
	"github.com/datatok/tintin/pkg/reporting/output"
)

const buildTemplateHelp = `
Generate the report_io, from Djobi jobs / stages log.
`

func newReportBuildAsTemplateCmd(client *action.ReportBuild, out io.Writer) *cobra.Command {

	cmd := &cobra.Command{
		Use:     "template",
		Aliases: []string{"html"},
		Short:   buildTemplateHelp,
		Long:    buildTemplateHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			report := client.Run()

			output.NewReportHTML(settings.ReportHTMLTemplatePath, report).ToHTML(out)

			return nil
		},
	}

	return cmd
}
