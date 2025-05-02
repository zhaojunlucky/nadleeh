package minio

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"                 // The MinIO Go client SDK
	"github.com/minio/minio-go/v7/pkg/credentials" // For providing credentials
	log "github.com/sirupsen/logrus"
	"nadleeh/pkg/env"
	"nadleeh/pkg/workflow/run_context"
	"os"
	"path/filepath"
)

type Minio struct {
	ctx       *run_context.WorkflowRunContext
	config    map[string]string
	URL       string
	AccessKey string
	SecretKey string
	Bucket    string
	Path      string
	Name      string
}

func (m *Minio) Init(ctx *run_context.WorkflowRunContext, config map[string]string) error {
	m.ctx = ctx
	m.config = config
	return nil
}

func (m *Minio) Run(parent env.Env) error {
	log.Infof("Run minio plugin")
	err := m.validate(parent)
	if err != nil {
		return err
	}

	minioClient, err := minio.New(m.URL, &minio.Options{
		Creds:  credentials.NewStaticV4(m.AccessKey, m.SecretKey, ""),
		Secure: true,
	})
	if err != nil {
		return fmt.Errorf("error initializing MinIO client: %v", err)
	}
	name := m.Name
	if len(name) == 0 {
		name = filepath.Base(m.Path)
	}
	file, err := os.Open(m.Path)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	fi, err := file.Stat()
	if err != nil {
		return fmt.Errorf("error getting file info: %v", err)
	}
	defer file.Close()
	info, err := minioClient.PutObject(context.Background(), m.Bucket, name, file, fi.Size(), minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("error uploading file: %v", err)
	}

	log.Infof("uploaded file: %s", info.Location)

	return nil

}

func (m *Minio) validate(parent env.Env) error {
	m.URL = parent.Expand(m.config["url"])
	if len(m.URL) <= 0 {
		return fmt.Errorf("invalid url")
	}
	m.AccessKey = parent.Expand(m.config["access-key"])
	if len(m.AccessKey) <= 0 {
		return fmt.Errorf("invalid access-key")
	}
	m.SecretKey = parent.Expand(m.config["secret-key"])
	if len(m.SecretKey) <= 0 {
		return fmt.Errorf("invalid secret-key")
	}
	m.Bucket = parent.Expand(m.config["bucket"])
	if len(m.Bucket) <= 0 {
		return fmt.Errorf("invalid bucket")
	}
	m.Path = parent.Expand(m.config["path"])

	m.Name = parent.Expand(m.config["name"])

	return nil
}
