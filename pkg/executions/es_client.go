package executions

import (
	"crypto/tls"
	"github.com/datatok/tintin/pkg/utils/cli"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

func getElasticsearchClient(settings *cli.EnvSettings) *elasticsearch.Client {
	es, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses:  strings.Split(settings.MetricsLogAPIURL, ","),
		MaxRetries: 2,
		Transport: &http.Transport{
			MaxIdleConnsPerHost:   10,
			ResponseHeaderTimeout: time.Second,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	})

	if err != nil {
		logrus.Warnf("nop, cannot connect to elasticsearch: %s", err)
		return nil
	}

	i, err := es.Info()

	if err != nil {
		logrus.Warnf("nop, cannot connect to elasticsearch: %s", err)
		return nil
	}

	logrus.Infof("connected: %s", i.String())

	return es
}
