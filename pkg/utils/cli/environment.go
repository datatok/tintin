package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/pflag"
)

// EnvSettings describes all of the environment settings.
type EnvSettings struct {
	// Djobi jobs & stages logs ES URL
	MetricsLogAPIURL string

	// Pipelines
	PipelinesURL, PipelinesPath string

	// Template
	ReportHTMLTemplatePath string

	// Logging stuff
	LogLevel string

	FrontURLPath string

	Debug bool
}

func New() *EnvSettings {

	env := EnvSettings{
		MetricsLogAPIURL:       envOr("METRICS_LOG_API_URL", "http://localhost:9200"),
		FrontURLPath:           envOr("FRONT_URLS_PATH", ""),
		PipelinesURL:           envOr("TINTIN_PIPELINES_URL", "."),
		PipelinesPath:          envOr("TINTIN_PIPELINES_PATH", "."),
		ReportHTMLTemplatePath: envOr("HTML_TEMPLATE", "./templates/index.html"),
		LogLevel:               envOr("LOG_LEVEL", "info"),
	}

	env.Debug, _ = strconv.ParseBool(os.Getenv("DEBUG"))

	return &env
}

// logrus.SetLevel(logrus.DebugLevel)

// AddFlags binds flags to the given flagset.
func (s *EnvSettings) AddFlags(fs *pflag.FlagSet) {
	fs.StringVarP(&s.ReportHTMLTemplatePath, "html_template", "", s.ReportHTMLTemplatePath, "Report HTML template path")
	fs.StringVarP(&s.FrontURLPath, "front_urls", "", s.FrontURLPath, "Path to YAML front linksRepository store")
	fs.StringVarP(&s.LogLevel, "log_level", "", s.LogLevel, "Log level (debug, info, warn, error)")
	fs.BoolVar(&s.Debug, "debug", s.Debug, "enable verbose output")
}

func envOr(name, def string) string {
	if v, ok := os.LookupEnv(name); ok {
		return v
	}
	return def
}

func (s *EnvSettings) EnvVars() map[string]string {
	envvars := map[string]string{
		"TINTIN_BIN":          os.Args[0],
		"DEBUG":               fmt.Sprint(s.Debug),
		"METRICS_LOG_API_URL": s.MetricsLogAPIURL,
		"PIPELINES_BUCKET":    s.PipelinesURL,
		"HTML_TEMPLATE":       s.ReportHTMLTemplatePath,
		"FRONT_URLS_PATH":     s.FrontURLPath,
		"LOG_LEVEL":           s.LogLevel,
	}

	return envvars
}
