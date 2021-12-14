package main

import (
	"bytes"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/datatok/tintin/pkg/action"
	"github.com/datatok/tintin/pkg/reporting/output"
	"github.com/datatok/tintin/pkg/reporting/sender"
)

const buildTemplateEmailHelp = `
Send report as HTML via email.
`

func newReportBuildAsEmailCmd(client *action.ReportBuild, out io.Writer) *cobra.Command {
	s := &sender.EmailSender{}
	e := sender.Email{}

	cmd := &cobra.Command{
		Use:   "email",
		Short: buildTemplateEmailHelp,
		Long:  buildTemplateEmailHelp,
		RunE: func(cmd *cobra.Command, args []string) error {
			report := client.Run()

			r := bytes.NewBufferString("")

			t := output.NewReportHTML(settings.ReportHTMLTemplatePath, report)

			t.ShowWorkLinks = false

			t.ToHTML(r)

			e.From = "@todo"
			e.Title = fmt.Sprintf("Djobi report %s", report.Title)
			e.Body = r.String()

			s.Send(e)

			return nil
		},
	}

	f := cmd.Flags()

	f.StringVar(&s.Server, "server", "", "SMTP server")
	f.StringVar(&e.To, "to", "", "Recipients, comma separated")

	return cmd
}
