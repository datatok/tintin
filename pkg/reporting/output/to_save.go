package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/datatok/tintin/pkg/utils"

	"github.com/Rican7/conjson"
	"github.com/Rican7/conjson/transform"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/sirupsen/logrus"

	"github.com/datatok/tintin/pkg/pipelines"
	"github.com/datatok/tintin/pkg/reporting"
)

type SaveStore struct {
	client    *elasticsearch.Client
	storeName string
}

/*
type DocumentAggregated struct {
	Counters PipelineCounters `json:"counters"`

	Title string
	Date string `json:"@timestamp"`
}
*/

type DocumentStorePipeline struct {
	Name, Team, File string
}

type DocumentStoreJob struct {
	ID, Name string
	Context  pipelines.JobContextDefinition
}

type DocumentStoreStageMetricValue struct {
	ValueStr, ValueDisplay, Unit string
	Value                        int
}

type DocumentStoreStageMetrics struct {
	Status, Reason, Link string
	Timeline             utils.ExecutionTimeline
	Value                DocumentStoreStageMetricValue
}

type DocumentStoreStage struct {
	URL, Name, Vendor, Version string
}

type DocumentStore struct {
	Pipeline DocumentStorePipeline
	Job      DocumentStoreJob
	Stage    DocumentStoreStage
	Metrics  DocumentStoreStageMetrics
	Date     string `json:"timestamp"`
}

func NewStore() *SaveStore {
	es, err := elasticsearch.NewDefaultClient()

	if err != nil {
		log.Fatalf("Error creating the client: %s", err)
	}

	return &SaveStore{
		es,
		"djobi-tintin",
	}
}

func (thisRepo *SaveStore) SaveReport(report *reporting.Report) {
	thisRepo.saveReportRaw(report)
}

/*
func (thisRepo *SaveStore) saveReportAggregated(report Report) {
	doc := DocumentAggregated{
		Counters: report.Counters,
		Title:    report.Title,
	}

	marshaler := conjson.NewMarshaler(doc, transform.ConventionalKeys())

	b, _ := marshaler.MarshalJSON()

	req := esapi.IndexRequest{
		Index:      thisRepo.aggregatedIndex,
		DocumentID: report.ID,
		Body:       strings.NewReader(string(b)),
		Refresh:    "true",
	}

	res, err := req.Do(context.Background(), thisRepo.client)

	if err != nil {
		log.Fatalf("Error getting response: %s", err)
	}

	defer res.Body.Close()

	if res.IsError() {
		log.Printf("[%s] Error indexing document ID=%d", res.Status(), report.ID)
	} else {
		// Deserialize the response into a map.
		var r map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&r); err != nil {
			log.Printf("Error parsing the response body: %s", err)
		} else {
			// Print the response Status and indexed document Version.
			log.Printf("[%s] %s; Version=%d", res.Status(), r["result"], int(r["_version"].(float64)))
		}
	}
}
*/
func (thisRepo *SaveStore) saveReportRaw(report *reporting.Report) {
	var (
		buf bytes.Buffer
	)

	i := 0
	batch := 50

	for _, pipeline := range report.Pipelines {
		for _, job := range pipeline.Jobs {
			for _, work := range job.Works {
				for stageID, stage := range work.Stages {
					doc := DocumentStore{
						Date: time.Now().Format(time.RFC3339),
						Pipeline: DocumentStorePipeline{
							Name: pipeline.Definition.Name,
							Team: pipeline.Definition.Team,
							File: pipeline.Definition.Path,
						},
						Job: DocumentStoreJob{
							ID:      work.Name,
							Name:    job.Name,
							Context: work.Context,
						},
						Stage: DocumentStoreStage{
							Vendor:  "Tintin",
							Version: "1.0.0",
							Name:    stageID,
						},
						Metrics: DocumentStoreStageMetrics{
							Status:   stage.Resume.Status,
							Reason:   stage.Resume.Details,
							Link:     stage.Resume.Link,
							Timeline: work.Timeline,
							Value: DocumentStoreStageMetricValue{
								ValueStr:     stage.Log.PostCheck.Meta.Count,
								ValueDisplay: stage.Log.PostCheck.Meta.Display,
								Unit:         stage.Log.PostCheck.Meta.Unit,
								Value:        stage.Log.PostCheck.Meta.Value,
							},
						},
					}

					// Prepare the metadata payload
					//
					meta := []byte(fmt.Sprintf(`{ "index" : { } }%s`, "\n"))

					// Prepare the data payload: encode article to JSON
					//
					marshaler := conjson.NewMarshaler(doc, transform.ConventionalKeys())

					data, err := marshaler.MarshalJSON()

					if err != nil {
						log.Fatalf("Cannot encode engine result: %s", err)
					}

					// Append newline to the data payload
					//
					data = append(data, "\n"...)

					// Append payloads to the buffer (ignoring write errors)
					//
					buf.Grow(len(meta) + len(data))
					buf.Write(meta)
					buf.Write(data)

					// When a threshold is reached, execute the Bulk() request with body from buffer
					//
					if i > 0 && i%batch == 0 {
						thisRepo.sendBulk(buf)

						buf.Reset()
					}

					i++
				}
			}
		}
	}

	if buf.Len() > 0 {
		thisRepo.sendBulk(buf)
	}

	logrus.Infof("Saved %d documents", i)

	buf.Reset()
}

/**
 *
type bulkResponse struct {
	Errors bool `json:"errors"`
	Items  []struct {
		Index struct {
			ID     string `json:"_id"`
			Result string `json:"result"`
			Status int    `json:"Status"`
			Error  struct {
				Type   string `json:"type"`
				Reason string `json:"Reason"`
				Cause  struct {
					Type   string `json:"type"`
					Reason string `json:"Reason"`
				} `json:"caused_by"`
			} `json:"error"`
		} `json:"index"`
	} `json:"items"`
}
*/

/**
 * Send ES bulk.
 */
func (thisRepo *SaveStore) sendBulk(bufferData bytes.Buffer) {
	var (
		raw map[string]interface{}
	)

	logrus.Debug("Sending bulk to ES")

	res, err := thisRepo.client.Bulk(bytes.NewReader(bufferData.Bytes()), thisRepo.client.Bulk.WithIndex(thisRepo.storeName))

	if err != nil {
		log.Fatalf("Failure indexing batch %s", err)
	}
	// If the whole request failed, print error and mark all documents as failed
	//
	if res.IsError() {
		if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
			log.Fatalf("Failure to to parse response body: %s", err)
		} else {
			log.Printf("  Error: [%d] %s: %s",
				res.StatusCode,
				raw["error"].(map[string]interface{})["type"],
				raw["error"].(map[string]interface{})["Reason"],
			)
		}
		// A successful response might still contain errors for particular documents...
		//
	} else {
		if err := json.NewDecoder(res.Body).Decode(&raw); err != nil {
			log.Fatalf("Failure to to parse response body: %s", err)
		} else {

		}
	}
}
