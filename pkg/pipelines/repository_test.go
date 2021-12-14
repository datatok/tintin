package pipelines

import (
	"fmt"

	"testing"

	"github.com/datatok/tintin/pkg/utils"
	"github.com/datatok/tintin/pkg/utils/cli"

	"github.com/stretchr/testify/assert"
)

func TestRepository_FindDefinitionsFromLocalDir(t *testing.T) {
	repo := NewRepository(&cli.EnvSettings{
		PipelinesPath: "./testdata/pipelines",
	})

	t.Run("must find 3 pipelines", func(t *testing.T) {
		pipelines, _ := repo.FindDefinitions(utils.Filter{})

		assert.Equal(t, 3, len(pipelines))
	})

	t.Run("must find 2 pipelines for team_a", func(t *testing.T) {
		pipelines, _ := repo.FindDefinitions(utils.Filter{
			Team: "team_a",
		})

		assert.Equal(t, 2, len(pipelines))
	})

	t.Run("analyze definitions without context", func(t *testing.T) {
		a := assert.New(t)

		pipelines, _ := repo.FindDefinitions(utils.Filter{
			Team: "team_a",
		})

		pipeline := pipelines[0]

		a.Equal("team_a", pipeline.Team)

		a.Equal(1, len(pipeline.Jobs))

		if a.Contains(pipeline.Jobs, "archivr") {
			job := pipeline.Jobs["archivr"]

			if a.NotNil(job) {
				a.Equal(2, len(job.Stages))
				a.Equal(1, len(job.Contexts))

				if a.Contains(job.Contexts, "_default_") {
					a.Equal(ContextTypeDefault, job.Contexts["_default_"].Type, "compare default context type")
				}
			}
		}
	})

	t.Run("analyze definitions with contexts", func(t *testing.T) {
		a := assert.New(t)

		pipelines, _ := repo.FindDefinitions(utils.Filter{Team: "team_a"})

		if a.Len(pipelines, 2) {
			pipeline := pipelines[1]

			a.Equal("team_a", pipeline.Team)

			a.Equal(1, len(pipeline.Jobs))

			if a.Contains(pipeline.Jobs, "conso") {
				job := pipeline.Jobs["conso"]

				if a.NotNil(job) {
					a.Equal(2, len(job.Stages))
					a.Equal(2, len(job.Contexts))

					if a.Contains(job.Contexts, "a") {
						a.Equal(ContextTypeSet, job.Contexts["a"].Type, "compare default context type")
					}
				}
			}
		}

		pipelines, _ = repo.FindDefinitions(utils.Filter{
			Team:      "team_b",
			Pipelines: "archivr",
		})

		if a.Len(pipelines, 1) {
			pipeline := pipelines[0]

			a.Equal("team_b/archivr", pipeline.FullName)

			if a.Contains(pipeline.Jobs, "archivr") {
				job := pipeline.Jobs["archivr"]

				if a.NotNil(job) {
					a.Equal(3, len(job.Stages))

					if a.Contains(job.Stages, "output") {
						stageOutput2 := job.Stages["output"]

						a.Equal(true, stageOutput2.IsEnabled())
					}

					if a.Contains(job.Stages, "output_2") {
						stageOutput2 := job.Stages["output_2"]

						a.Equal(false, stageOutput2.IsEnabled())
					}
				}
			}
		}
	})
}

/*
func TestRepository_FindDefinitions(t *testing.T) {
	repo := NewRepository(&cli.EnvSettings{
		PipelinesURL: "steam",
	})

	t.Run("must find more than 3 pipelines", func(t *testing.T) {
		pipelines := repo.FindDefinitions(utils.Filter{})

		assert.Greater(t, len(pipelines), 3)
	})

	t.Run("must find 1 pipelines for admin", func(t *testing.T) {
		pipelines := repo.FindDefinitions(utils.Filter{
			Team: "admin",
		})

		assert.Equal(t, 1, len(pipelines))
	})
}*/

func TestStageDefinition_IsOutputStage(t *testing.T) {
	tests := []struct {
		definition StageDefinition
		isOutput   bool
	}{
		{
			definition: StageDefinition{
				Name:  "",
				Stage: "",
				Kind:  "",
			},
			isOutput: false,
		},
		{
			definition: StageDefinition{
				Name:  "org.es.output",
				Stage: "output",
				Kind:  "",
			},
			isOutput: true,
		},
		{
			definition: StageDefinition{
				Name:  "hive",
				Stage: "org.es.output",
				Kind:  "",
			},
			isOutput: true,
		},
		{
			definition: StageDefinition{
				Name:  "gnegne",
				Stage: "output",
				Kind:  "",
			},
			isOutput: true,
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			if tt.definition.IsOutputStage() != tt.isOutput {
				t.Errorf("expected %t, got %t for definition %s", tt.isOutput, tt.definition.IsOutputStage(), tt.definition.Name)
			}
		})
	}
}
