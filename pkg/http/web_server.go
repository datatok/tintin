package http

import (
	"encoding/json"

	"fmt"

	"github.com/sirupsen/logrus"

	"github.com/datatok/tintin/pkg/executions"

	"io"

	"net/http"
	"net/url"

	"github.com/datatok/tintin/pkg/engine"
	"github.com/datatok/tintin/pkg/metrics"
	"github.com/datatok/tintin/pkg/pipelines"
	"github.com/datatok/tintin/pkg/reporting/output"
	"github.com/datatok/tintin/pkg/utils"
	"github.com/datatok/tintin/pkg/utils/cli"

	"strings"
	"time"
)

type WebServer struct {
	Port     int
	settings *cli.EnvSettings
}

func NewWebServer(s *cli.EnvSettings) *WebServer {
	return &WebServer{settings: s}
}

func (thisWebServer *WebServer) Run(out io.Writer) error {
	http.HandleFunc("/status", thisWebServer.GetStatus)
	http.HandleFunc("/", thisWebServer.HelloServer)
	http.Handle("/favicon.ico", http.FileServer(http.Dir("./web")))
	http.Handle("/metrics", metrics.New(thisWebServer.settings).HTTPEndpoint())

	fmt.Printf("Starting web server, on port %d\n\n", thisWebServer.Port)

	err := http.ListenAndServe(fmt.Sprintf(":%d", thisWebServer.Port), nil)

	if err != nil {
		logrus.Fatal(err)
	}

	return nil
}

func (thisWebServer *WebServer) GetStatus(out http.ResponseWriter, r *http.Request) {
	repo := pipelines.NewRepository(thisWebServer.settings)

	ret := map[string]string{
		"web_server":            "ok: because you see this...",
		"elasticsearch_metrics": executions.GetServiceStatus(thisWebServer.settings),
		"email_server":          "todo (i am really lazy)",
		"storage_s3":            repo.GetStorageStatus(),
	}

	out.Header().Add("Content-Type", "application/json")
	out.WriteHeader(200)

	retStr, _ := json.Marshal(ret)

	_, _ = out.Write(retStr)
}

func (thisWebServer *WebServer) HelloServer(out http.ResponseWriter, r *http.Request) {
	argTeam := getOrDefault(r.URL, "team", "")
	argPipeline := getOrDefault(r.URL, "pipeline", "*")
	argLevel := getOrDefault(r.URL, "status", "")
	argSchedule := getOrDefault(r.URL, "date",
		time.Now().AddDate(0, 0, -1).Format("02/01/2006"))
	argLevels := make([]string, 0)

	if len(argLevel) > 0 {
		argLevels = strings.Split(argLevel, ",")
	}
	filter := utils.Filter{
		Schedule:  argSchedule,
		Team:      argTeam,
		Pipelines: argPipeline,
		Status:    argLevels,
	}

	repo := pipelines.NewRepository(thisWebServer.settings)
	definitions, err := repo.FindDefinitions(filter)

	if err == nil {
		checker := engine.New(thisWebServer.settings, filter)

		rp := checker.Execute(definitions)

		if len(argLevels) > 0 {
			rp = rp.FilterByLevel(argLevels)
		}

		out.Header().Add("Content-Type", "text/html")
		out.WriteHeader(200)

		t := output.NewReportHTML(thisWebServer.settings.ReportHTMLTemplatePath, rp)

		t.ToHTML(out)
	} else {
		logrus.Error(err)

		out.WriteHeader(500)
		out.Write([]byte(err.Error()))
	}
}

func getOrDefault(u *url.URL, key string, def string) string {
	v := u.Query().Get(key)

	if len(v) == 0 {
		return def
	}

	return v
}
