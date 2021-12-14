package executions

import (
	"encoding/json"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/sirupsen/logrus"

	"github.com/datatok/tintin/pkg/pipelines"
	"github.com/datatok/tintin/pkg/utils"
	"github.com/datatok/tintin/pkg/utils/cli"
)

type SparkExecutor struct {
	Spark struct {
		Spark struct {
			Application struct {
				Name, ID string
			}
		}
	}
}

type JobExecution struct {
	PreCheckStatus  string `json:"pre_check_status"`
	RunStatus       string `json:"run_status"`
	PostCheckStatus string `json:"post_check_status"`
	Executor        SparkExecutor
	Timeline        utils.ExecutionTimeline
	Pipeline        *struct {
		UID, Name  string
		Definition pipelines.Definition
	}
	Args map[string]string
	UID  string
	ID   string
}

type JobSearchAPIResponse struct {
	Took int
	Hits struct {
		Total struct {
			Value int
		}
		Hits []struct {
			ID         string          `json:"_id"`
			Source     JobExecution    `json:"_source"`
			Highlights json.RawMessage `json:"highlight"`
			Sort       []interface{}   `json:"sort"`
		}
	}
}

type JobsStore struct {
	client              *elasticsearch.Client
	StoreName           string
	filterScheduleTitle string
	JobExecutions       []JobExecution
}

func GetServiceStatus(settings *cli.EnvSettings) string {
	es := getElasticsearchClient(settings)
	i, err := es.Info()

	if err == nil {
		return i.String()
	}

	return "error: " + err.Error()
}

func NewJobsStore(settings *cli.EnvSettings, scheduleTitle string) *JobsStore {
	es := getElasticsearchClient(settings)

	return &JobsStore{
		client:              es,
		StoreName:           "djobi-jobs",
		filterScheduleTitle: scheduleTitle,
	}
}

// Fetch djobi-jobs
//
func (c *JobsStore) FetchJobsExecutions() {
	var (
		r JobSearchAPIResponse
	)

	client := c.client

	res, err := client.Search(
		client.Search.WithIndex(c.StoreName),
		client.Search.WithQuery("meta.title.keyword:"+strings.Replace(c.filterScheduleTitle, "/", "\\/", -1)),
		client.Search.WithSize(1000),
	)

	if err != nil {
		logrus.Errorf("Error getting response: %s", err)
	} else {
		defer res.Body.Close()

		if res.IsError() {
			var e map[string]interface{}
			if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
				logrus.Errorf("Error parsing the response body: %s", err)
			} else {
				// Print the response status and error information.
				logrus.Errorf("[%s] %s: %s",
					res.Status(),
					e["error"].(map[string]interface{})["type"],
					e["error"].(map[string]interface{})["reason"],
				)
			}
		} else {
			if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
				logrus.Errorf("Error parsing the response body: %s", err)
			}

			for _, hit := range r.Hits.Hits {
				c.JobExecutions = append(c.JobExecutions, hit.Source)
			}
		}
	}
}

/**
 * Get the pipeline jobs execution
 */
func (c *JobsStore) FindJobExecution(pipeline pipelines.Definition, id string) *JobExecution {

	for _, jobExecution := range c.JobExecutions {
		if (jobExecution.Pipeline.Name == pipeline.FullName || strings.HasSuffix(pipeline.FullName, jobExecution.Pipeline.Name)) && jobExecution.ID == id {
			jobExecution.Pipeline.Definition = pipeline

			return &jobExecution
		}
	}

	return nil
}
