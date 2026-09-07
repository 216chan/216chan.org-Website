package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type Storage interface {
	Save(ctx context.Context, filename string, data io.Reader, size int64, contentType string) (string, error)
	Delete(ctx context.Context, url string) error
}

func New(uploadDir string) Storage {
	bucket := os.Getenv("S3_BUCKET")
	if bucket != "" {
		region := os.Getenv("S3_REGION")
		if region == "" {
			region = "us-east-1"
		}
		endpoint := os.Getenv("S3_ENDPOINT") // for R2: https://xxx.r2.cloudflarestorage.com
		return newS3(bucket, region, endpoint)
	}
	return newLocal(uploadDir)
}


type localFS struct {
	dir string
	urlPrefix string
}

func newLocal(dir string) *localFS {
	if dir == "" {
		dir = "./uploads"
		if _, err := os.Stat("/data"); err == nil {
			dir = "/data/uploads"
		}
	}
	_ = os.MkdirAll(dir, 0755)
	return &localFS{dir: dir, urlPrefix: "/uploads"}
}

func (l *localFS) Save(_ context.Context, filename string, data io.Reader, _ int64, _ string) (string, error) {
	safe := filepath.Base(filename)
	if safe == "" || safe == "." {
		return "", fmt.Errorf("invalid filename")
	}
	dst := filepath.Join(l.dir, safe)
	f, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, data); err != nil {
		return "", err
	}
	return l.urlPrefix + "/" + safe, nil
}

func (l *localFS) Delete(_ context.Context, url string) error {
	if !strings.HasPrefix(url, l.urlPrefix+"/") {
		return nil
	}
	name := strings.TrimPrefix(url, l.urlPrefix+"/")
	name = filepath.Base(name)
	return os.Remove(filepath.Join(l.dir, name))
}


type s3Storage struct {
	bucket   string
	client   *s3.Client
	publicBase string
}

func newS3(bucket, region, endpoint string) *s3Storage {
	cfg, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(region))
	if err != nil {
		return &s3Storage{bucket: bucket}
	}
	var client *s3.Client
	if endpoint != "" {
		client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(endpoint)
			o.UsePathStyle = true
		})
	} else {
		client = s3.NewFromConfig(cfg)
	}
	publicBase := os.Getenv("S3_PUBLIC_URL") // e.g. https://cdn.216chan.org
	return &s3Storage{bucket: bucket, client: client, publicBase: publicBase}
}

func (s *s3Storage) Save(ctx context.Context, filename string, data io.Reader, size int64, contentType string) (string, error) {
	if s.client == nil {
		l := newLocal("")
		return l.Save(ctx, filename, data, size, contentType)
	}
	safe := filepath.Base(filename)
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(safe),
		Body:          data,
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	})
	if err != nil {
		return "", err
	}
	if s.publicBase != "" {
		return strings.TrimSuffix(s.publicBase, "/") + "/" + safe, nil
	}
	return fmt.Sprintf("https://%s.s3.amazonaws.com/%s", s.bucket, safe), nil
}

func (s *s3Storage) Delete(ctx context.Context, url string) error {
	if s.client == nil {
		return nil
	}
	parts := strings.Split(url, "/")
	key := parts[len(parts)-1]
	if key == "" {
		return nil
	}
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}