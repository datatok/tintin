package cli

import (
	"os"
	"strings"
	"testing"

	"github.com/spf13/pflag"
)

func TestEnvSettings(t *testing.T) {
	tests := []struct {
		name string

		// input
		args   string
		envars map[string]string

		// expected values
		es    string
		debug bool
	}{
		{
			name: "defaults",
			es:   "http://localhost:9200",
		},
		{
			name:  "with flags set",
			args:  "--debug --metrics_log_url=http://localhost:9200",
			es:    "http://localhost:9200",
			debug: true,
		},
		{
			name:   "with envvars set",
			envars: map[string]string{"DEBUG": "1", "METRICS_LOG_API_URL": "http://localhost:9200"},
			es:     "http://localhost:9200",
			debug:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer resetEnv()()

			for k, v := range tt.envars {
				os.Setenv(k, v)
			}

			flags := pflag.NewFlagSet("testing", pflag.ContinueOnError)

			settings := New()
			settings.AddFlags(flags)
			flags.Parse(strings.Split(tt.args, " "))

			if settings.Debug != tt.debug {
				t.Errorf("expected debug %t, got %t", tt.debug, settings.Debug)
			}
			if settings.MetricsLogAPIURL != tt.es {
				t.Errorf("expected ES URL %q, got %q", tt.es, settings.MetricsLogAPIURL)
			}
		})
	}
}

func resetEnv() func() {
	origEnv := os.Environ()

	// ensure any local envvars do not hose us
	for e := range New().EnvVars() {
		os.Unsetenv(e)
	}

	return func() {
		for _, pair := range origEnv {
			kv := strings.SplitN(pair, "=", 2)
			os.Setenv(kv[0], kv[1])
		}
	}
}
