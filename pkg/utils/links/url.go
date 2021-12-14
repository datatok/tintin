package links

import (
	"bytes"
	"html/template"
	"io/ioutil"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

type Repository struct {
	Path      string
	Templates map[string]*template.Template
}

/**
 * Load YAML file
 */
func Load(path string) Repository {
	if len(path) == 0 {
		return Repository{
			Path: path,
		}
	}

	dat, err := ioutil.ReadFile(path)

	if err != nil {
		logrus.Fatal(err)
	}

	return LoadFromString(string(dat))
}

/**
 * Load YAML file
 */
func LoadFromString(str string) Repository {
	ret := Repository{
		Path: "inline",
	}

	var entries map[string]string

	err := yaml.Unmarshal([]byte(str), &entries)

	if err != nil {
		logrus.Fatal(err)
	}

	ret.Templates = make(map[string]*template.Template)

	for entryName, entryValue := range entries {
		tmpl, errT := template.New("reporting").Parse(entryValue)

		if errT == nil {
			ret.Templates[entryName] = tmpl
		} else {
			logrus.Error(errT)
		}
	}

	return ret
}

/**
 * Resolve URL
 */
func (r Repository) Generate(name string, data map[string]string) string {
	if tpl, ok := r.Templates[name]; ok {
		var buffer bytes.Buffer

		if err := tpl.Execute(&buffer, data); err != nil {
			logrus.Error(err)

			return "__ERROR__:" + err.Error()
		}

		return buffer.String()
	}

	return "__NOT_FOUND__:" + name
}
