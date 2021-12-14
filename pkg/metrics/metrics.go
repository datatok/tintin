package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/push"
	"github.com/sirupsen/logrus"

	"net/http"

	"github.com/datatok/tintin/pkg/engine"
	"github.com/datatok/tintin/pkg/pipelines"
	"github.com/datatok/tintin/pkg/utils"
	"github.com/datatok/tintin/pkg/utils/cli"

	"time"
)

type Metrics struct {
	registry *prometheus.Registry
	settings *cli.EnvSettings
}

var (
	stagesProcessed = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "djobi_works_total",
		Help: "The total number of pipeline stages, per team and pipeline.",
	},
		[]string{
			"team",
			"pipeline",
			"pipeline_fullname",
		})

	worksDuration = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "djobi_work_duration",
		Help: "Work duration, in ms.",
	},
		[]string{
			"team",
			"pipeline",
			"pipeline_fullname",
			"djobi_job",
			"djobi_work",
		})

	worksExecutionDetails = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "djobi_work_execution_data_count",
		Help: ".",
	},
		[]string{
			"team",
			"pipeline",
			"pipeline_fullname",
			"djobi_job",
			"djobi_work",
			"djobi_stage",
		})
)

func New(settings *cli.EnvSettings) *Metrics {
	r := prometheus.NewRegistry()

	r.MustRegister(stagesProcessed, worksDuration, worksExecutionDetails)

	return &Metrics{
		registry: r,
		settings: settings,
	}
}

func (metrics *Metrics) buildMetrics() {
	filter := utils.Filter{
		Schedule:  time.Now().AddDate(0, 0, -1).Format("02/01/2006"),
		Team:      "",
		Pipelines: "",
		Status:    make([]string, 0),
	}

	repo := pipelines.NewRepository(metrics.settings)
	definitions, err := repo.FindDefinitions(filter)

	if err == nil {
		checker := engine.New(metrics.settings, filter)

		rp := checker.Execute(definitions)

		for _, p := range rp.Pipelines {
			stagesProcessed.WithLabelValues(p.Definition.Team, p.Definition.Name, p.Definition.FullName).Set(float64(p.Counters.Works))

			for _, j := range p.Jobs {
				for _, w := range j.Works {
					worksDuration.
						WithLabelValues(p.Definition.Team, p.Definition.Name, p.Definition.FullName, j.Name, w.Name).
						Set(float64(w.Timeline.Duration))

					for _, s := range w.Stages {
						worksExecutionDetails.
							WithLabelValues(p.Definition.Team, p.Definition.Name, p.Definition.FullName, j.Name, w.Name, s.Log.Kind).
							Set(float64(s.Log.PostCheck.Meta.Value))
					}
				}
			}
		}
	} else {
		logrus.Error(err)
	}
}

func (metrics *Metrics) Push(to string) {

	metrics.buildMetrics()

	err := push.New(to, "djobi-tintin").Gatherer(metrics.registry).Push()

	if err != nil {
		logrus.Fatalf("Error sending data: %s", err)
	}
}

func (metrics *Metrics) HTTPEndpoint() http.Handler {
	promHandler := promhttp.HandlerFor(metrics.registry, promhttp.HandlerOpts{})

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		metrics.buildMetrics()

		promHandler.ServeHTTP(w, r)
	})
}
