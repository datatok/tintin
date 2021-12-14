package pipelines

import (
	"io/ioutil"
	"regexp"
	"strings"

	"github.com/datatok/tintin/pkg/utils"
	"github.com/datatok/tintin/pkg/utils/cli"
	"github.com/datatok/tintin/pkg/utils/links"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const (
	ContextTypeSet     = "set"
	ContextTypeDefault = "default"
	GitlabURL          = "pipelines_gitlab"
)

type repository interface {
	getDefinitions() ([]Definition, error)
}

type StageDefinition struct {
	Name, Stage, Enabled, Kind string
}

type JobContextDefinition struct {
	Name, Type string
}

type JobDefinition struct {
	Name     string
	Stages   map[string]StageDefinition
	Contexts map[string]JobContextDefinition
}

type MetaOwnerDefinition struct {
	Name, Email, Role string
}

type MetaDefinition struct {
	Team   string
	Owners []MetaOwnerDefinition
}

type ReportingDefinition struct {
	Enabled bool `default:true`
}

/**
 * Repository pipeline definition (from YAML file).
 */
type Definition struct {
	Path, FullName, Name, Team, GitlabLink string

	Jobs map[string]JobDefinition

	Meta MetaDefinition

	Reporting ReportingDefinition
}

type Repository struct {
	URL, Path       string
	linksRepository links.Repository
}

func NewRepository(settings *cli.EnvSettings) *Repository {
	return &Repository{
		URL:             settings.PipelinesURL,
		Path:            settings.PipelinesPath,
		linksRepository: links.Load(settings.FrontURLPath),
	}
}

/**
 * Find pipeline definitions, from S3 service.
 */
func (s *Repository) FindDefinitions(filter utils.Filter) ([]Definition, error) {

	var client repository

	if len(s.URL) == 0 {
		client = &RepositoryGit{
			path:            s.Path,
			linksRepository: s.linksRepository,
		}
	} else if strings.HasPrefix(s.URL, "s3://") {
		client = &RepositoryS3{
			bucket:          s.Path,
			linksRepository: s.linksRepository,
		}
	} else {
		file, err := ioutil.TempDir("/tmp", "tintin")

		if err != nil {
			logrus.Fatal(err)
		}

		client = &RepositoryGit{
			URL:             s.URL,
			path:            s.Path,
			workingDir:      file,
			linksRepository: s.linksRepository,
		}
	}

	definitions, err := client.getDefinitions()

	logrus.Infof("Found %d pipelines", len(definitions))

	definitions = s.filterDefinitions(definitions, filter)

	logrus.Infof("After filter: %d pipelines", len(definitions))

	return definitions, err
}

func (s *Repository) filterDefinitions(definitions []Definition, filter utils.Filter) []Definition {
	var (
		ret               []Definition
		filterPipelineReg *regexp.Regexp
	)

	if len(filter.Pipelines) > 0 && filter.Pipelines != "*" {
		filterPipelineReg = regexp.MustCompile(filter.Pipelines)
	}

	for _, definition := range definitions {
		if filterPipelineReg != nil && !filterPipelineReg.Match([]byte(definition.Path)) {
			continue
		}

		if len(filter.Team) > 0 && filter.Team != definition.Team {
			continue
		}

		if !definition.Reporting.Enabled {
			continue
		}

		ret = append(ret, definition)
	}

	return ret
}

func (s *Repository) GetStorageStatus() string {
	return "ok"
}

/**
 * Is stage enabled
 */
func (stage StageDefinition) IsEnabled() bool {
	return stage.Enabled != "false"
}

/**
 * Is output stage
 */
func (stage StageDefinition) IsOutputStage() bool {
	name := stage.Stage
	kind := stage.Kind

	return strings.Contains(kind, "-output") ||
		strings.Contains(kind, ".output") ||
		kind == "output" ||
		name == "output" ||
		strings.Contains(name, "output")
}

/**
 * Read content
 */
func (def *Definition) parsePipeline(pipelineContent []byte) {
	err := yaml.Unmarshal(pipelineContent, def)

	for jobName, job := range def.Jobs {
		job.Name = jobName
		if job.Contexts == nil || len(job.Contexts) == 0 {
			job.Contexts = make(map[string]JobContextDefinition)
			job.Contexts["_default_"] = JobContextDefinition{"_default_", ContextTypeDefault}
		} else {
			for contextName, context := range job.Contexts {
				context.Name = contextName
				context.Type = ContextTypeSet

				job.Contexts[contextName] = context
			}
		}

		def.Jobs[jobName] = job
	}

	check(err)
}

func defaultPipelineDefinition(path string, fullName string, name string, team string, gitlabLink string) Definition {
	return Definition{
		Path:       path,
		FullName:   fullName,
		Name:       name,
		Team:       team,
		GitlabLink: gitlabLink,
		Jobs:       nil,
		Meta: MetaDefinition{
			Team: team,
		},
		Reporting: ReportingDefinition{
			Enabled: true,
		},
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
