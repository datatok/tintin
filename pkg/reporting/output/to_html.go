package output

import (
	"bytes"
	"fmt"

	"html/template"
	"io"
	"io/ioutil"
	"runtime"
	"strconv"
	"strings"
	gotime "time"

	"github.com/datatok/tintin/internal/version"

	"github.com/sirupsen/logrus"
	"github.com/ulule/deepcopier"

	"github.com/datatok/tintin/pkg/reporting"
	"github.com/datatok/tintin/pkg/utils/constant"

	"os"
)

type ReportHTML struct {
	TemplatePath  string
	ShowWorkLinks bool
	Report        *reporting.Report
}

func NewReportHTML(templatePath string, report *reporting.Report) *ReportHTML {
	return &ReportHTML{
		TemplatePath:  templatePath,
		Report:        report,
		ShowWorkLinks: true,
	}
}

func (rHTML *ReportHTML) ToHTML(out io.Writer) {
	file, err := os.Open(rHTML.TemplatePath)

	if file != nil {
		defer file.Close()
	}

	if err != nil {
		logrus.Error(err)
	} else {
		html, _ := ioutil.ReadAll(file)

		htmlAsStr := strings.Replace(string(html), "\t", "", -1)

		tmpl, errT := template.New("reporting").Funcs(map[string]interface{}{
			"report_url":     rHTML.Report.Link.Build,
			"duration":       ParseDuration,
			"link_to":        rHTML.LinkTo,
			"date":           ParseDate,
			"time":           ParseTime,
			"job_color":      jobColor,
			"pipeline_color": pipelineColor,
			"nl2br":          Nl2Br,
			"percentage":     Percentage,
		}).Parse(htmlAsStr)

		if errT != nil {
			logrus.Error(errT)
		}

		bufferOut := bytes.NewBufferString("")

		err = tmpl.Execute(bufferOut, map[string]interface{}{
			"report":          rHTML.Report,
			"Counters":        rHTML.Report.Counters,
			"show_work_links": rHTML.ShowWorkLinks,
			"BuildInfo":       version.Get(),
			"RuntimeVersion":  runtime.Version(),
		})

		lines := strings.Split(bufferOut.String(), "\n")

		for _, line := range lines {
			if len(strings.TrimSpace(line)) > 0 {
				l := strings.TrimLeft(line, " ")

				out.Write([]byte(l))

				if len(l) > 15 {
					out.Write([]byte("\n"))
				}
			}
		}

		if err != nil {
			panic(err)
		}
	}
}

// ParseDuration -
func ParseDuration(n int) gotime.Duration {
	d, _ := gotime.ParseDuration(strconv.Itoa(n) + "ms")

	d = d.Round(gotime.Second)

	return d
}

func (rHTML *ReportHTML) LinkTo(k string, v string) string {

	cloneLink := &reporting.ReportLink{}

	_ = deepcopier.Copy(&rHTML.Report.Link).To(cloneLink)

	switch k {
	case "date":
		cloneLink.Arguments.Date = v
	case "team":
		cloneLink.Arguments.Team = v
	case "pipeline":
		cloneLink.Arguments.Pipeline = v
	case "status":
		cloneLink.Arguments.Status = []string{v}
	}

	return cloneLink.Build()
}

// Since -
func Since(n gotime.Time) gotime.Duration {
	return gotime.Since(n)
}

// Since -
func ParseDate(d string) string {
	dd, err := gotime.Parse("2006-01-02T15:04:05.000+0000", d)

	if err != nil {
		logrus.Warn(err)
		return d
	}

	return dd.Format(gotime.Stamp)
}

func ParseTime(d string) string {
	dd, err := gotime.Parse("2006-01-02T15:04:05.000+0000", d)

	if err != nil {
		logrus.Warn(err)
		return d
	}

	return dd.Format(gotime.Kitchen)
}

func pipelineColor(pipeline reporting.Pipeline) string {

	for _, job := range pipeline.Jobs {
		for _, work := range job.Works {
			if work.Status == constant.DoneError {
				return "danger"
			}
		}
	}

	for _, job := range pipeline.Jobs {
		for _, work := range job.Works {
			if !work.Success {
				return "warning"
			}
		}
	}

	return "success"
}

func jobColor(job reporting.Job) string {
	for _, work := range job.Works {
		if work.Status == constant.DoneError {
			return "danger"
		}
	}

	for _, work := range job.Works {
		if !work.Success {
			return "warning"
		}
	}

	return "success"
}

func Percentage(a int, b int) string {
	if b == 0 {
		return "0%"
	}

	return fmt.Sprintf("%d %s", int(100*a/b), "%")
}

// Nl2Br is breakstr inserted before looks like space (CRLF , LFCR, SPACE, NL)
func Nl2Br(str string) template.HTML {

	// BenchmarkNl2Br-8                   	10000000	      3398 ns/op
	// BenchmarkNl2BrUseStringReplace-8   	10000000	      4535 ns/op
	brtag := []byte("<br />")
	l := len(str)
	buf := make([]byte, 0, l) //prealloca

	for i := 0; i < l; i++ {

		switch str[i] {

		case 10, 13: //NL or CR

			buf = append(buf, brtag...)

			if l >= i+1 {
				if l > i+1 && (str[i+1] == 10 || str[i+1] == 13) { //NL+CR or CR+NL
					i++
				}
			}
		default:
			buf = append(buf, str[i])
		}
	}

	return template.HTML(string(buf))
}
