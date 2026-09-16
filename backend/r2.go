package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type iconStorage interface {
	Put(context.Context, string, io.Reader, string) error
	Delete(context.Context, string) error
	PublicURL(string) string
}

type r2Storage struct {
	client        *s3.Client
	bucket        string
	publicBaseURL string
}

func newR2Storage(ctx context.Context) (*r2Storage, error) {
	endpoint := strings.TrimRight(strings.TrimSpace(os.Getenv("R2_ENDPOINT")), "/")
	bucket := strings.Trim(strings.TrimSpace(os.Getenv("R2_BUCKET_NAME")), "/")
	accessKeyID := strings.TrimSpace(os.Getenv("R2_ACCESS_KEY_ID"))
	secretAccessKey := strings.TrimSpace(os.Getenv("R2_SECRET_ACCESS_KEY"))
	publicBaseURL := strings.TrimRight(strings.TrimSpace(os.Getenv("R2_PUBLIC_BASE_URL")), "/")

	missing := make([]string, 0, 5)
	for name, value := range map[string]string{
		"R2_ENDPOINT":          endpoint,
		"R2_BUCKET_NAME":       bucket,
		"R2_ACCESS_KEY_ID":     accessKeyID,
		"R2_SECRET_ACCESS_KEY": secretAccessKey,
		"R2_PUBLIC_BASE_URL":   publicBaseURL,
	} {
		if value == "" {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("required R2 environment variables are missing: %s", strings.Join(missing, ", "))
	}

	// Cloudflare sometimes displays an endpoint with the bucket appended. The
	// SDK takes the endpoint and bucket separately.
	endpoint = strings.TrimSuffix(endpoint, "/"+bucket)
	if err := validateHTTPURL(endpoint, "R2_ENDPOINT"); err != nil {
		return nil, err
	}
	if err := validateHTTPURL(publicBaseURL, "R2_PUBLIC_BASE_URL"); err != nil {
		return nil, err
	}
	publicURL, _ := url.Parse(publicBaseURL)
	if strings.HasSuffix(publicURL.Hostname(), ".r2.cloudflarestorage.com") {
		return nil, errors.New("R2_PUBLIC_BASE_URL must be an r2.dev URL or custom public domain, not the S3 API endpoint")
	}

	awsConfig, err := config.LoadDefaultConfig(
		ctx,
		config.WithRegion("auto"),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, ""),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating R2 configuration: %w", err)
	}
	client := s3.NewFromConfig(awsConfig, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(endpoint)
		options.UsePathStyle = true
	})
	return &r2Storage{client: client, bucket: bucket, publicBaseURL: publicBaseURL}, nil
}

func (storage *r2Storage) Put(ctx context.Context, key string, body io.Reader, contentType string) error {
	_, err := storage.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(storage.bucket),
		Key:          aws.String(key),
		Body:         body,
		ContentType:  aws.String(contentType),
		CacheControl: aws.String("public, max-age=31536000, immutable"),
	})
	if err != nil {
		return fmt.Errorf("uploading icon to R2: %w", err)
	}
	return nil
}

func (storage *r2Storage) Delete(ctx context.Context, key string) error {
	if strings.TrimSpace(key) == "" {
		return nil
	}
	_, err := storage.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(storage.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("deleting icon from R2: %w", err)
	}
	return nil
}

func (storage *r2Storage) PublicURL(key string) string {
	return storage.publicBaseURL + "/" + strings.TrimLeft(key, "/")
}

func validateHTTPURL(value, name string) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" ||
		(parsed.Scheme != "http" && parsed.Scheme != "https") ||
		parsed.RawQuery != "" || parsed.Fragment != "" {
		return errors.New(name + " must be an HTTP(S) URL without a query or fragment")
	}
	return nil
}
