package pipelines

import (
	"fmt"
	"github.com/datatok/tintin/pkg/utils/links"
	"github.com/go-git/go-git/v5"
	"github.com/sirupsen/logrus"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type RepositoryGit struct {
	URL, path, workingDir string
	linksRepository       links.Repository
}

func (s *RepositoryGit) getDefinitions() ([]Definition, error) {

	var (
		ret            []Definition
		candidatesPath []string
	)

	if len(s.URL) > 0 {
		logrus.Debugf("Cloning repo into %s", s.workingDir)

		_, err := git.PlainClone(s.workingDir, false, &git.CloneOptions{
			URL:      s.URL,
			Progress: os.Stdout,
		})

		if err != nil {
			logrus.Fatal(err)
		}
	}

	searchPath := filepath.Join(s.workingDir, s.path)
	workingDirSanitized := sanitizePath(searchPath)

	filepath.Walk(searchPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return nil
		}

		if !info.IsDir() && (filepath.Base(path) == "pipeline.yml" || filepath.Base(path) == "pipeline.yaml") {
			candidatesPath = append(candidatesPath, path)
		}

		return nil
	})

	for _, candidatePath := range candidatesPath {
		candidatePathSanitized := sanitizePath(candidatePath)
		pp := strings.Replace(candidatePathSanitized, workingDirSanitized, "", 1)
		fullName := strings.Trim(filepath.Dir(strings.Replace(candidatePathSanitized, workingDirSanitized, "", 1)), "/")
		team := "steam"
		name := fullName

		if strings.Contains(fullName, "/") {
			team = strings.Split(fullName, "/")[0]
			name = strings.Trim(strings.TrimLeft(fullName, team), "/")
		}

		def := defaultPipelineDefinition(candidatePath, fullName, name, team, s.linksRepository.Generate(GitlabURL, map[string]string{"uri": pp}))

		dat, err := os.ReadFile(candidatePath)

		if err != nil {
			logrus.Fatal(err)
		}

		def.parsePipeline(dat)

		ret = append(ret, def)
	}

	return ret, nil
}

func sanitizePath(p string) string {
	p = strings.TrimLeft(p, "./")
	p = strings.TrimLeft(p, "../")

	return p
}
