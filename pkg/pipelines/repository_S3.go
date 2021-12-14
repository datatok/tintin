package pipelines

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/datatok/tintin/pkg/utils/links"
	"github.com/sirupsen/logrus"
)

type RepositoryS3 struct {
	client          *s3.S3
	bucket          string
	linksRepository links.Repository
}

func getS3Client() *s3.S3 {
	mySession := session.Must(session.NewSession(&aws.Config{
		Region:           aws.String(os.Getenv("AWS_REGION")),
		Endpoint:         aws.String(os.Getenv("S3_URL")),
		S3ForcePathStyle: aws.Bool(true),
	}))

	return s3.New(mySession)
}

func (s *RepositoryS3) GetStorageStatus() string {
	svc := getS3Client()

	keyPrefix := "datahub/develop/pipelines"

	_, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(keyPrefix),
	})

	if err == nil {
		return "ok, S3 connected, URL found"
	}

	return "nop: " + err.Error()
}

/**
 * Find pipeline definitions, from S3 service.
 */
func (s *RepositoryS3) getDefinitions() ([]Definition, error) {
	keyPrefix := "datahub/develop/pipelines"

	svc := getS3Client()

	logrus.WithField("URL", s.bucket).Infof("Searching pipelines")

	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(keyPrefix),
	})

	if err != nil {
		return nil, fmt.Errorf("Unable to list items in URL %q, %v", s.bucket, err)
	}

	var (
		ret []Definition
	)

	for _, item := range resp.Contents {
		value := *item.Key
		fullName := strings.Trim(filepath.Dir(strings.Replace(value, keyPrefix, "", 1)), "/")

		if !strings.Contains(fullName, "/dev") &&
			(strings.HasSuffix(value, ".yml") || strings.HasSuffix(value, ".yaml")) {
			pp := strings.Replace(value, keyPrefix, "", 1)
			team := "steam"
			name := fullName

			if strings.Contains(fullName, "/") {
				team = strings.Split(fullName, "/")[0]
				name = strings.Trim(strings.TrimLeft(fullName, team), "/")
			}

			def := defaultPipelineDefinition(value, fullName, name, team, s.linksRepository.Generate(GitlabURL, map[string]string{"uri": pp}))

			rawObject, err := svc.GetObject(
				&s3.GetObjectInput{
					Bucket: aws.String(s.bucket),
					Key:    aws.String(value),
				})
			if err != nil {
				return nil, fmt.Errorf("Unable to download item %q, %v", item, err)
			}

			buf := new(bytes.Buffer)
			buf.ReadFrom(rawObject.Body)

			def.parsePipeline(buf.Bytes())

			ret = append(ret, def)
		}
	}

	return ret, nil
}
