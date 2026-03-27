package storage

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Config holds the configuration for the S3-compatible storage backend.
// Set Endpoint and UsePathStyle for MinIO or other non-AWS providers.
type S3Config struct {
	Bucket       string
	Region       string
	Endpoint     string
	AccessKey    string
	SecretKey    string
	UsePathStyle bool
}

type S3Storage struct {
	client       *s3.Client
	bucket       string
	region       string
	endpoint     string
	usePathStyle bool
}

// NewS3 creates an S3Storage client. If AccessKey/SecretKey are empty,
// credentials are loaded from the default AWS credential chain.
func NewS3(ctx context.Context, cfg S3Config) (*S3Storage, error) {
	loadOptions := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(cfg.Region),
	}

	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		loadOptions = append(loadOptions, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		))
	}

	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = cfg.UsePathStyle
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
	})

	return &S3Storage{
		client:       client,
		bucket:       cfg.Bucket,
		region:       cfg.Region,
		endpoint:     cfg.Endpoint,
		usePathStyle: cfg.UsePathStyle,
	}, nil
}

func (s *S3Storage) Upload(ctx context.Context, key string, reader io.Reader) (string, error) {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
		Body:   reader,
	})
	if err != nil {
		return "", err
	}
	return s.objectURL(key), nil
}

func (s *S3Storage) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (s *S3Storage) objectURL(key string) string {
	key = strings.TrimLeft(key, "/")
	if s.endpoint != "" {
		base := strings.TrimRight(s.endpoint, "/")
		if s.usePathStyle {
			return fmt.Sprintf("%s/%s/%s", base, s.bucket, key)
		}
		return fmt.Sprintf("%s/%s", base, key)
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, key)
}
