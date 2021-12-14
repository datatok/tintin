package links

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name, content, url string
	}{
		{
			name: "Normal",
			url:  "__NOT_FOUND__:es_conso",
			content: `
toto: https://es:9200{{ .uri }}
`,
		},
		{
			name: "Normal",
			url:  "http://es:9200/_search?q=",
			content: `
es_conso: http://es:9200{{ .uri }}
toto: https://es:9200{{ .uri }}
`,
		},
	}

	a := assert.New(t)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a.Equal(tt.url, LoadFromString(tt.content).Generate("es_conso", map[string]string{"uri": "/_search?q="}))
		})
	}
}
