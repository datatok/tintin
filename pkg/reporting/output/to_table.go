package output

import (
	"io"
	"strings"

	"github.com/olekukonko/tablewriter"

	"github.com/datatok/tintin/pkg/reporting"
)

func ToTable(out io.Writer, report *reporting.Report) {
	data := [][]string{}

	table := tablewriter.NewWriter(out)
	table.SetHeader([]string{"Pipeline", "Job", "Contexts"})

	for _, pipeline := range report.Pipelines {

		for _, job := range pipeline.Jobs {

			var (
				color    int    = tablewriter.FgGreenColor
				contexts string = ""
			)

			for _, c := range job.Works {
				contexts = contexts + ", " + c.Context.Name + " (" + c.Details + ")"

				if !c.Success {
					color = tablewriter.FgRedColor
				}
			}

			contexts = strings.Trim(strings.Trim(contexts, ", "), " ")

			table.Rich([]string{
				pipeline.Definition.Team + " > " + pipeline.Definition.Name,
				job.Name,
				contexts,
			}, []tablewriter.Colors{{}, {}, {color}})

		}
	}

	for _, v := range data {
		table.Append(v)
	}
	table.Render() // Send output
}
