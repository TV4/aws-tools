package main

import (
	"context"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	awsclient "github.com/aws/aws-sdk-go/aws/client"
	awscreds "github.com/aws/aws-sdk-go/aws/credentials"
	awssession "github.com/aws/aws-sdk-go/aws/session"
	awss3 "github.com/aws/aws-sdk-go/service/s3"
	awss3manager "github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3Client is a client for interacting with an S3 bucket.
type S3Client struct {
	region      string
	httpTimeout time.Duration

	accessKeyID     string
	secretAccessKey string

	awsConfProviderMu sync.Mutex
	awsConfProvider   awsclient.ConfigProvider
}

// NewS3Client creates a new S3 client.
func NewS3Client(region string, opts ...func(*S3Client)) *S3Client {
	c := &S3Client{
		region:      region,
		httpTimeout: 2 * time.Hour,
	}

	for _, o := range opts {
		o(c)
	}

	return c
}

// WithCredentials is an option to configure client credentials when
// instantiating a new client.
func WithCredentials(accessKeyID, secretAccessKey string) func(*S3Client) {
	return func(c *S3Client) {
		c.accessKeyID = accessKeyID
		c.secretAccessKey = secretAccessKey
	}
}

// List returns a list of all keys in the given bucket.
func (c *S3Client) List(ctx context.Context, bucket, prefix string) ([]string, error) {
	cfgProvider, err := c.configProvider()
	if err != nil {
		return nil, err
	}

	s3c := awss3.New(cfgProvider)

	var keys []string

	loi := &awss3.ListObjectsInput{
		Prefix: aws.String(prefix),
		Bucket: aws.String(bucket),
	}

	for {
		res, err := s3c.ListObjects(loi)
		if err != nil {
			return nil, err
		}

		for _, obj := range res.Contents {
			keys = append(keys, *obj.Key)
		}

		if !*res.IsTruncated {
			break
		}

		loi.Marker = aws.String(keys[len(keys)-1])
	}

	return keys, nil
}

// Open opens the given object for reading.
func (c *S3Client) Open(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	cfgProvider, err := c.configProvider()
	if err != nil {
		return nil, err
	}

	s3c := awss3.New(cfgProvider)

	goi := &awss3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}

	res, err := s3c.GetObjectWithContext(ctx, goi)
	if err != nil {
		return nil, err
	}

	return res.Body, nil
}

// Delete deletes the object at the given S3 key.
func (c *S3Client) Delete(ctx context.Context, bucket, key string) error {
	cfgProvider, err := c.configProvider()
	if err != nil {
		return err
	}

	s3c := awss3.New(cfgProvider)

	doi := &awss3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &key,
	}

	if _, err := s3c.DeleteObjectWithContext(ctx, doi); err != nil {
		return err
	}

	return nil
}

// Upload reads data from src and uploads it to the given S3 key.
func (c *S3Client) Upload(ctx context.Context, bucket, key string, src io.Reader) error {
	cfgProvider, err := c.configProvider()
	if err != nil {
		return err
	}

	uploader := awss3manager.NewUploader(cfgProvider)

	uploadInput := &awss3manager.UploadInput{
		Body:   src,
		Bucket: &bucket,
		Key:    &key,
	}

	if _, err := uploader.UploadWithContext(ctx, uploadInput); err != nil {
		return err
	}

	return nil
}

func (c *S3Client) configProvider() (awsclient.ConfigProvider, error) {
	c.awsConfProviderMu.Lock()
	defer c.awsConfProviderMu.Unlock()

	if c.awsConfProvider == nil {
		var creds *awscreds.Credentials

		if c.accessKeyID != "" && c.secretAccessKey != "" {
			creds = awscreds.NewStaticCredentials(c.accessKeyID, c.secretAccessKey, "")
		}

		sess, err := awssession.NewSession(
			&aws.Config{
				Credentials: creds,
				Region:      &c.region,
				HTTPClient: &http.Client{
					Timeout: c.httpTimeout,
				},
			},
		)
		if err != nil {
			return nil, err
		}

		c.awsConfProvider = sess
	}

	return c.awsConfProvider, nil
}
