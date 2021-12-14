package executions

import (
	"encoding/json"
	"log"

	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"

	"github.com/datatok/tintin/pkg/utils"
	"github.com/datatok/tintin/pkg/utils/cli"
)

type Meta struct {
	Value                                            int
	Count, Display, Reason, Unit, Query, Index, Size string
}

type StagePhase struct {
	Status string
	Link   string
	Meta   Meta
}

type StageHit struct {
	Status, Name, Stage string
	Kind                string `json:"type"`

	Job *struct {
		ID, UID string
	}

	PreCheck  StagePhase `json:"pre_check"`
	PostCheck StagePhase `json:"post_check"`

	Timeline utils.ExecutionTimeline

	Error *struct {
		Message string
	}
}

type StageSearchAPIResponse struct {
	Took int
	Hits struct {
		Total struct {
			Value int
		}
		Hits []struct {
			ID         string          `json:"_id"`
			Source     StageHit        `json:"_source"`
			Highlights json.RawMessage `json:"highlight"`
			Sort       []interface{}   `json:"sort"`
		}
	}
}

type StagesStore struct {
	client              *elasticsearch.Client
	StoreName           string
	filterScheduleTitle string
	JobExecutions       []JobExecution
}

func NewStagesStore(settings *cli.EnvSettings, scheduleTitle string) *StagesStore {
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: strings.Split(settings.MetricsLogAPIURL, ","),
	})

	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	return &StagesStore{
		client:              es,
		StoreName:           "djobi-stages",
		filterScheduleTitle: scheduleTitle,
	}
}

// Fetch djobi-stages
//

func (c *StagesStore) FetchStagesExecutions(jobExecutionUID string) []StageHit {
	var (
		r   StageSearchAPIResponse
		ret []StageHit
	)

	client := c.client

	res, err := client.Search(
		client.Search.WithIndex(c.StoreName),
		client.Search.WithQuery("job.uid:"+jobExecutionUID),
		client.Search.WithSize(1000),
	)

	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}

	defer res.Body.Close()

	if res.IsError() {
		var e map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&e); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		} else {
			// Print the response status and error information.
			log.Fatalf("[%s] %s: %s",
				res.Status(),
				e["error"].(map[string]interface{})["type"],
				e["error"].(map[string]interface{})["reason"],
			)
		}
	} else {
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Fatalf("Error parsing the response body: %s", err)
		}

		for _, hit := range r.Hits.Hits {
			ret = append(ret, hit.Source)
		}
	}

	// Fix legacy
	for i := range ret {
		stage := &ret[i]
		stage.PostCheck.Meta.Value = fixMetaCount(&stage.PostCheck)
	}

	return ret
}

/**
 * To support legacy logs.
 */
func fixMetaCount(phase *StagePhase) int {
	if len(phase.Meta.Size) > 0 {
		phase.Meta.Display = phase.Meta.Size

		buffer := strings.Split(phase.Meta.Size, " ")

		phase.Meta.Count = buffer[0]

		if len(buffer) == 2 {
			phase.Meta.Unit = buffer[1]
		}
	}

	if phase.Meta.Value == 0 && len(phase.Meta.Count) > 0 {
		i, err := strconv.Atoi(phase.Meta.Count)

		if err == nil {
			return i
		}
	}

	return phase.Meta.Value
}
