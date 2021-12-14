package reporting

import (
	"fmt"
	"log"
	"net/url"

	"strings"

	"github.com/datatok/tintin/pkg/utils/constant"

	"github.com/google/uuid"
	"github.com/ulule/deepcopier"

	"github.com/datatok/tintin/pkg/executions"
	"github.com/datatok/tintin/pkg/pipelines"
	"github.com/datatok/tintin/pkg/utils"
)

const (
	PublicFrontReportURL = "report_front"
)

type Status struct {
	Status  string
	Details string
	Link    string
}

type WorkStageDetails struct {
	Log    executions.StageHit
	Resume Status

	PreCheck  Status `json:"pre_check"`
	Run       Status
	PostCheck Status `json:"post_check"`
}

type Work struct {
	Context pipelines.JobContextDefinition

	Stages map[string]WorkStageDetails

	Timeline utils.ExecutionTimeline

	Success bool

	Name, Status, Details, Link, LinkToJobLogs, LinkToJobStagesLogs, LinkToSparkHistory, LinkToYARNHistory string
}

type Job struct {
	ID    string
	Name  string
	Works []Work
}

type PipelineCounters struct {
	Jobs, Works, Contexts, Success, Unknown, Errors, Executions int
}

type Pipeline struct {
	UID string `json:"uid"`

	Jobs []Job

	Counters PipelineCounters

	Definition pipelines.Definition
}

type ReportLink struct {
	URL       string
	Arguments ReportLinkArguments
}

type ReportLinkArguments struct {
	Team, Date, Pipeline string
	Status               []string
}

type Report struct {
	ID, Title string
	Link      ReportLink
	Filter    utils.Filter
	Counters  PipelineCounters
	Pipelines []Pipeline
}

/*
 * Build full URL
 */
func (link *ReportLink) Build() string {
	v := url.Values{}

	if len(link.Arguments.Date) > 0 {
		v.Set("date", link.Arguments.Date)
	}

	if len(link.Arguments.Team) > 0 {
		v.Set("team", link.Arguments.Team)
	}

	if len(link.Arguments.Pipeline) > 0 {
		v.Set("pipeline", link.Arguments.Pipeline)
	}

	if len(link.Arguments.Status) > 0 {
		v.Set("status", link.Arguments.Status[0])
	}

	return link.URL + "?" + v.Encode()
}

func NewReport(filter utils.Filter) *Report {
	docID, err := uuid.NewUUID()

	if err != nil {
		log.Fatalf("Error building ID: %s", err)
	}

	return &Report{
		ID:     docID.String(),
		Filter: filter,
	}
}

func (r *Report) CalculateCounters() {
	for _, p := range r.Pipelines {
		r.Counters.Jobs += len(p.Jobs)

		for _, j := range p.Jobs {
			r.Counters.Contexts += len(j.Works)

			for _, w := range j.Works {
				if w.Status == constant.DoneOk {
					r.Counters.Success++
				} else if w.Status == constant.DoneError {
					r.Counters.Errors++
				} else if w.Status == constant.DoneUnknown {
					r.Counters.Unknown++
				}

				if w.Status != constant.No {
					r.Counters.Executions++
				}

			}
		}
	}

	r.Title = fmt.Sprintf("Djobi report for %s - %d success / %d errors / %d unknowns",
		r.Filter.Schedule,
		r.Counters.Success,
		r.Counters.Errors,
		r.Counters.Unknown,
	)
}

/**
 * Filter work by status (success , error ...).
 */
func (r *Report) FilterByLevel(levels []string) *Report {
	newReport := &Report{}

	// Copy report (counters, title ...)
	_ = deepcopier.Copy(r).To(newReport)

	// Reset pipelines list
	newReport.Pipelines = []Pipeline{}

	levelsMap := fixLevels(levels)

	for _, pipeline := range r.Pipelines {
		copyPipeline := Pipeline{
			UID:        pipeline.UID,
			Definition: pipeline.Definition,
			Counters:   PipelineCounters{},
		}
		for _, job := range pipeline.Jobs {
			copyJob := Job{
				ID:   job.ID,
				Name: job.Name,
			}
			for _, work := range job.Works {
				if _, ok := levelsMap[work.Status]; ok {
					copyJob.Works = append(copyJob.Works, work)
					copyPipeline.Counters.Works++
				}
			}

			if len(copyJob.Works) > 0 {
				copyPipeline.Jobs = append(copyPipeline.Jobs, copyJob)
				copyPipeline.Counters.Jobs++
			}
		}

		if len(copyPipeline.Jobs) > 0 {
			newReport.Pipelines = append(newReport.Pipelines, copyPipeline)
		}
	}

	//ret.CalculateCounters()

	return newReport
}

func fixLevels(levels []string) map[string]string {
	ret := make(map[string]string)

	for _, level := range levels {
		level = strings.ToUpper(level)

		if level == "ERROR" {
			level = constant.DoneError
		} else if level == "OK" || level == "SUCCESS" {
			level = constant.DoneOk
		} else if level == "WARNING" || level == "WARN" {
			level = constant.DoneUnknown
		}

		ret[level] = level
	}

	return ret
}
