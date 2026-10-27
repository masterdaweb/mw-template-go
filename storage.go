package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// storage is the private bucket bound to this service: the variables the
// binding injects, prefix MW_STORAGE_. Its files are never public; each link
// is signed and expires.
type storage struct {
	bucket  string
	client  *s3.Client
	presign *s3.PresignClient
}

type check struct {
	OK     bool   `json:"ok"`
	Detail string `json:"detail"`
}

type file struct {
	Key      string    `json:"key"`
	Name     string    `json:"name"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
	URL      string    `json:"url"`
}

func newStorage() (*storage, error) {
	bucket := os.Getenv("MW_STORAGE_BUCKET")
	if bucket == "" {
		return &storage{}, errors.New("MW_STORAGE_BUCKET is empty: connect a bucket to this service")
	}
	region := os.Getenv("MW_STORAGE_REGION")
	if region == "" {
		region = "auto"
	}
	client := s3.New(s3.Options{
		Region:       region,
		BaseEndpoint: aws.String(os.Getenv("MW_STORAGE_ENDPOINT")),
		UsePathStyle: os.Getenv("MW_STORAGE_FORCE_PATH_STYLE") == "true",
		Credentials: credentials.NewStaticCredentialsProvider(
			os.Getenv("MW_STORAGE_ACCESS_KEY_ID"), os.Getenv("MW_STORAGE_SECRET_ACCESS_KEY"), ""),
	})
	return &storage{bucket: bucket, client: client, presign: s3.NewPresignClient(client)}, nil
}

func (s *storage) check(ctx context.Context) check {
	if s.client == nil {
		return check{Detail: "MW_STORAGE_BUCKET is empty: connect a bucket to this service"}
	}
	if _, err := s.client.HeadBucket(ctx, &s3.HeadBucketInput{Bucket: &s.bucket}); err != nil {
		return check{Detail: firstLine(err.Error())}
	}
	return check{OK: true, Detail: s.bucket}
}

var unsafeName = regexp.MustCompile(`[^\w.-]+`)

func (s *storage) put(ctx context.Context, name, contentType string, body io.Reader, size int64) (file, error) {
	if s.client == nil {
		return file{}, errors.New("no bucket connected")
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	clean := unsafeName.ReplaceAllString(path.Base(name), "-")
	key := fmt.Sprintf("files/%s/%s", uuid.NewString(), clean)
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: &s.bucket, Key: &key, Body: body, ContentLength: &size, ContentType: &contentType,
	})
	if err != nil {
		return file{}, err
	}
	link, err := s.link(ctx, key, clean)
	return file{Key: key, Name: clean, Size: size, Modified: time.Now(), URL: link}, err
}

func (s *storage) list(ctx context.Context) ([]file, error) {
	if s.client == nil {
		return nil, errors.New("no bucket connected")
	}
	out, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: &s.bucket, Prefix: aws.String("files/")})
	if err != nil {
		return nil, err
	}
	files := make([]file, 0, len(out.Contents))
	for _, o := range out.Contents {
		name := path.Base(*o.Key)
		link, err := s.link(ctx, *o.Key, name)
		if err != nil {
			return nil, err
		}
		files = append(files, file{Key: *o.Key, Name: name, Size: aws.ToInt64(o.Size), Modified: aws.ToTime(o.LastModified), URL: link})
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Modified.After(files[j].Modified) })
	if len(files) > 20 {
		files = files[:20]
	}
	return files, nil
}

// link is a download link good for 15 minutes.
func (s *storage) link(ctx context.Context, key, name string) (string, error) {
	req, err := s.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket:                     &s.bucket,
		Key:                        &key,
		ResponseContentDisposition: aws.String(fmt.Sprintf(`attachment; filename="%s"`, strings.ReplaceAll(name, `"`, ""))),
	}, s3.WithPresignExpires(15*time.Minute))
	if err != nil {
		return "", err
	}
	return req.URL, nil
}

func firstLine(s string) string {
	s, _, _ = strings.Cut(s, "\n")
	if len(s) > 160 {
		s = s[:160]
	}
	return s
}
