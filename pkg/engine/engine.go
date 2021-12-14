package engine

import (
	"fmt"

	"github.com/dustin/go-humanize"
	"github.com/sirupsen/logrus"

	"github.com/datatok/tintin/pkg/executions"
	"github.com/datatok/tintin/pkg/pipelines"
	"github.com/datatok/tintin/pkg/reporting"
	"github.com/datatok/tintin/pkg/utils"
	"github.com/datatok/tintin/pkg/utils/cli"
	"github.com/datatok/tintin/pkg/utils/constant"
	"github.com/datatok/tintin/pkg/utils/links"

	"strings"
)

const (
	MetricsLogServerFrontURL = "djobi_es_search"
	SparkHistoryFrontURL     = "spark_history"
	YARNHistoryFrontURL      = "yarn_history"
)

type Checker struct {
	settings *cli.EnvSettings
	jobs     *executions.JobsStore
	stages   *executions.StagesStore
	filter   utils.Filter
	urls     links.Repository
}

func New(settings *cli.EnvSettings, filter utils.Filter) *Checker {
	return &Checker{
		settings: settings,
		filter:   filter,
		jobs:     executions.NewJobsStore(settings, filter.Schedule),
		stages:   executions.NewStagesStore(settings, filter.Schedule),
		urls:     links.Load(settings.FrontURLPath),
	}
}

/**
 * Generate the full report.
 */
func (c *Checker) Execute(pipelines []pipelines.Definition) *reporting.Report {
	rp := reporting.NewReport(c.filter)

	c.jobs.FetchJobsExecutions()

	for _, pipeline := range pipelines {
		rp.Pipelines = append(rp.Pipelines, c.Check(pipeline))
	}

	rp.Link = reporting.ReportLink{
		URL: c.urls.Generate(reporting.PublicFrontReportURL, map[string]string{}),
		Arguments: reporting.ReportLinkArguments{
			Date:     c.filter.Schedule,
			Status:   c.filter.Status,
			Team:     c.filter.Team,
			Pipeline: c.filter.Pipelines,
		},
	}

	rp.CalculateCounters()

	return rp
}

/**
 * Check the pipeline
 */
func (c *Checker) Check(pipeline pipelines.Definition) reporting.Pipeline {
	var (
		jobs     []reporting.Job
		counters reporting.PipelineCounters
	)

	for jobName, job := range pipeline.Jobs {
		reportJob := reporting.Job{
			Name: jobName,
		}

		for _, contextDefinition := range job.Contexts {
			res := c.checkWork(pipeline, job, contextDefinition)

			reportJob.Works = append(reportJob.Works, res)
		}

		jobs = append(jobs, reportJob)

		counters.Works += len(reportJob.Works)
	}

	return reporting.Pipeline{
		UID:        "",
		Jobs:       jobs,
		Definition: pipeline,
		Counters:   counters,
	}
}

/**
 * Check piece of work : job context, on all stages.
 */
//nolint:ineffassign
func (c *Checker) checkWork(pipeline pipelines.Definition, job pipelines.JobDefinition, contextDefinition pipelines.JobContextDefinition) reporting.Work {
	var (
		stageExecutions []executions.StageHit
	)

	ret := reporting.Work{
		Context: contextDefinition,
		Name:    job.Name,
		Stages:  make(map[string]reporting.WorkStageDetails),
	}

	displayMessage := "No execution log found!"

	if contextDefinition.Type == pipelines.ContextTypeSet {
		ret.Name += "_" + contextDefinition.Name
	}

	logrus.Debugf("checking pipeline %s job %s", pipeline.Name, ret.Name)

	// Get djobi-jobs execution, for this pipeline execution
	jobExecution := c.jobs.FindJobExecution(pipeline, ret.Name)

	// If we found job execution -> find jobs stages executions
	if jobExecution != nil {
		ret.Timeline = jobExecution.Timeline

		ret.LinkToJobLogs = c.urls.Generate(MetricsLogServerFrontURL, map[string]string{"index": c.jobs.StoreName, "query": "_id:" + jobExecution.UID})
		ret.LinkToJobStagesLogs = c.urls.Generate(MetricsLogServerFrontURL, map[string]string{"index": c.stages.StoreName, "query": "job.uid:" + jobExecution.UID})
		ret.LinkToSparkHistory = c.urls.Generate(SparkHistoryFrontURL, map[string]string{"app_id": jobExecution.Executor.Spark.Spark.Application.ID})
		ret.LinkToYARNHistory = c.urls.Generate(YARNHistoryFrontURL, map[string]string{"app_id": jobExecution.Executor.Spark.Spark.Application.ID})

		stageExecutions = c.stages.FetchStagesExecutions(jobExecution.UID)

		if len(stageExecutions) == 0 {
			displayMessage = "Stage execution log is not found!"
		} else {
			displayMessage = ""
		}
	}

	// Loop definition stages => output stages first
	for _, stageDefinition := range job.Stages {
		if stageDefinition.IsEnabled() {
			stageExecution := searchStageExecution(stageDefinition, stageExecutions)

			// If pre-check in error
			if stageExecution != nil && stageExecution.PreCheck.Status == constant.DoneError {
				ret.Stages[fixStageName(stageDefinition.Kind)] = reporting.WorkStageDetails{
					Resume: reporting.Status{
						Status:  constant.DoneError,
						Details: stageExecution.PreCheck.Meta.Reason,
					},
				}
			}

			// If output stage OR stage has failed
			if stageDefinition.IsOutputStage() || (stageExecution != nil && stageExecution.Status == constant.DoneError) {
				if stageExecution == nil {
					ret.Stages[fixStageName(stageDefinition.Kind)] = reporting.WorkStageDetails{
						Resume: reporting.Status{
							Status:  constant.No,
							Details: "Stage execution is not found!",
						},
					}
				} else {
					ret.Stages[fixStageName(stageExecution.Kind)] = reporting.WorkStageDetails{
						Resume: c.stagePhaseToReportStatus(*stageExecution),
						Log:    *stageExecution,
					}
				}
			}
		}
	}

	// Loop definition stages => all enabled stages
	if len(ret.Stages) == 0 {
		for _, stageDefinition := range job.Stages {
			if stageDefinition.IsEnabled() {
				stageExecution := searchStageExecution(stageDefinition, stageExecutions)

				// If output stage OR stage has failed
				if stageExecution != nil {
					ret.Stages[fixStageName(stageExecution.Kind)] = reporting.WorkStageDetails{
						Resume: c.stagePhaseToReportStatus(*stageExecution),
						Log:    *stageExecution,
					}
				}
			}
		}
	}

	var (
		stageExecutionSuccess bool
		outStatus             string
	)

	if len(ret.Stages) == 0 {
		logrus.Warnf("stages not found for %s / %s", pipeline.FullName, job.Name)
	} else {
		stageExecutionSuccess = true
		outStatus = constant.DoneOk

		for _, s := range ret.Stages {
			if s.Log.Status == constant.DoneOk {
				if s.Log.PostCheck.Status == constant.DoneError {
					stageExecutionSuccess = false
					outStatus = constant.DoneError
				} else if s.Log.PostCheck.Status == constant.Todo ||
					s.Log.PostCheck.Status == constant.DoneUnknown ||
					s.Log.PostCheck.Status == constant.InProgress {
					stageExecutionSuccess = false
					outStatus = constant.DoneUnknown
				}
			} else {
				stageExecutionSuccess = false
				outStatus = constant.DoneError
			}
		}
	}

	fillStatus(&ret, stageExecutionSuccess, outStatus, displayMessage)

	return ret
}

/**
 * Search stage execution log
 */
func searchStageExecution(stageDefinition pipelines.StageDefinition, hits []executions.StageHit) *executions.StageHit {
	if len(hits) == 0 {
		return nil
	}

	for _, hit := range hits {
		if hit.Kind == stageDefinition.Kind {
			return &hit
		}
	}

	return nil
}

func fillStatus(work *reporting.Work, success bool, status string, reason string) {
	work.Success = success
	work.Status = status
	work.Details = reason
}

/**
 * If old stage definition
 */
func fixStageName(stageKind string) string {
	if strings.Contains(stageKind, ".") {
		return stageKind
	}

	switch stageKind {
	case "elasticsearch":
		return "org.elasticsearch.output"
	}

	return stageKind
}

func (c *Checker) stagePhaseToReportStatus(stage executions.StageHit) reporting.Status {

	phase := stage.PostCheck
	details := phase.Meta.Reason
	link := phase.Link

	switch phase.Status {
	case constant.DoneError:
		if phase.Meta.Value == 0 && strings.Contains(stage.Kind, "elasticsearch") {
			details = fmt.Sprintf("no document in %s", phase.Meta.Index)
			link = c.urls.Generate("es_conso", map[string]string{"index": phase.Meta.Index, "query": phase.Meta.Query})
		} else if strings.Contains(stage.Kind, "scp") {
			details = SCPBuildDisplay(stage)
		} else if len(phase.Meta.Reason) > 0 {
			details = phase.Meta.Reason
		} else {
			details = phase.Meta.Display
		}

	case constant.DoneOk:
		if strings.Contains(stage.Kind, "elasticsearch") {
			details = fmt.Sprintf("%s documents in %s", humanize.FormatInteger("# ###,", phase.Meta.Value), phase.Meta.Index)
			link = c.urls.Generate("es_conso", map[string]string{"index": phase.Meta.Index, "query": phase.Meta.Query})
		} else if strings.Contains(stage.Kind, "scp") {
			details = SCPBuildDisplay(stage) // fmt.Sprintf("File of %s", phase.Meta.Display)
		} else if len(phase.Meta.Reason) > 0 {
			details = phase.Meta.Reason
		} else if len(phase.Meta.Unit) > 0 {
			if phase.Meta.Unit == "byte" {
				details = ByteCountSI(int64(phase.Meta.Value))
			} else {
				details = fmt.Sprintf("%s %s", humanize.FormatInteger("# ###,", phase.Meta.Value), phase.Meta.Unit)
			}
		} else {
			details = phase.Meta.Display
		}

	case constant.Todo, constant.No, constant.DoneUnknown:
		if stage.PreCheck.Status == constant.DoneError {
			details = fmt.Sprintf("Pre check error: %s", stage.PreCheck.Meta.Reason)
		} else {
			details = "No execution"

			if stage.Status == constant.DoneOk {
				details += " (but run was ok)"
			} else {
				details += fmt.Sprintf(" (and run status is %s)", stage.Status)
			}
		}
	}

	if stage.Error != nil && stage.Error.Message != "" {
		details += "\nRun error: \"" + stage.Error.Message + "\""
	}

	return reporting.Status{
		Status:  phase.Status,
		Details: details,
		Link:    link,
	}
}

func ByteCountSI(b int64) string {
	const unit = 1000
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB",
		float64(b)/float64(div), "kMGTPE"[exp])
}
