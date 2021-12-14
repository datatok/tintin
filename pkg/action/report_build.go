package action

import (
	"github.com/datatok/tintin/pkg/engine"
	"github.com/datatok/tintin/pkg/pipelines"
	"github.com/datatok/tintin/pkg/reporting"
	"github.com/datatok/tintin/pkg/utils"
	"github.com/datatok/tintin/pkg/utils/cli"
	"github.com/sirupsen/logrus"
)

type ReportBuild struct {
	Filter   utils.Filter
	settings *cli.EnvSettings
}

func NewBuildReport(s *cli.EnvSettings) *ReportBuild {
	return &ReportBuild{
		settings: s,
	}
}

func (p *ReportBuild) Run() *reporting.Report {
	pp, err := pipelines.NewRepository(p.settings).FindDefinitions(p.Filter)

	if err != nil {
		logrus.Fatal(err)
	}

	checker := engine.New(p.settings, p.Filter)
	r := checker.Execute(pp)

	if len(p.Filter.Status) > 0 {
		r = r.FilterByLevel(p.Filter.Status)
	}

	return r
}
