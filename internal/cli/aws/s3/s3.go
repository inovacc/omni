// Package s3 provides AWS S3 operations
package s3

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	awscommon "github.com/inovacc/omni/internal/cli/aws"
)

// Client wraps the S3 client
type Client struct {
	client  *s3.Client
	presign *s3.PresignClient
	printer *awscommon.Printer
}

// NewClient creates a new S3 client
func NewClient(cfg aws.Config, w io.Writer, format awscommon.OutputFormat) *Client {
	client := s3.NewFromConfig(cfg)
	return &Client{
		client:  client,
		presign: s3.NewPresignClient(client),
		printer: awscommon.NewPrinter(w, format),
	}
}

// S3URI represents a parsed S3 URI
type S3URI struct {
	Bucket string
	Key    string
	IsS3   bool
}

// ParseS3URI parses an S3 URI (s3://bucket/key)
func ParseS3URI(uri string) (*S3URI, error) {
	if !strings.HasPrefix(uri, "s3://") {
		return &S3URI{IsS3: false}, nil
	}

	parsed, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("invalid S3 URI: %w", err)
	}

	key := strings.TrimPrefix(parsed.Path, "/")
	return &S3URI{
		Bucket: parsed.Host,
		Key:    key,
		IsS3:   true,
	}, nil
}

// LsOptions configures the ls operation
type LsOptions struct {
	Recursive bool
	Human     bool
	Summarize bool
}

// Object represents an S3 object
type Object struct {
	Key          string    `json:"Key"`
	LastModified time.Time `json:"LastModified"`
	Size         int64     `json:"Size"`
	StorageClass string    `json:"StorageClass,omitempty"`
	ETag         string    `json:"ETag,omitempty"`
}

// Bucket represents an S3 bucket
type Bucket struct {
	Name         string    `json:"Name"`
	CreationDate time.Time `json:"CreationDate"`
}

// Ls lists buckets or objects
func (c *Client) Ls(ctx context.Context, uri string, opts LsOptions) error {
	if uri == "" || uri == "s3://" {
		// List all buckets
		return c.listBuckets(ctx)
	}

	s3uri, err := ParseS3URI(uri)
	if err != nil {
		return err
	}

	if !s3uri.IsS3 {
		return fmt.Errorf("invalid S3 URI: %s", uri)
	}

	return c.listObjects(ctx, s3uri.Bucket, s3uri.Key, opts)
}

func (c *Client) listBuckets(ctx context.Context) error {
	result, err := c.client.ListBuckets(ctx, &s3.ListBucketsInput{})
	if err != nil {
		return fmt.Errorf("list-buckets: %w", err)
	}

	var buckets []Bucket
	for _, b := range result.Buckets {
		buckets = append(buckets, Bucket{
			Name:         aws.ToString(b.Name),
			CreationDate: aws.ToTime(b.CreationDate),
		})
	}

	return c.printer.PrintJSON(buckets)
}

func (c *Client) listObjects(ctx context.Context, bucket, prefix string, opts LsOptions) error {
	delimiter := "/"
	if opts.Recursive {
		delimiter = ""
	}

	paginator := s3.NewListObjectsV2Paginator(c.client, &s3.ListObjectsV2Input{
		Bucket:    aws.String(bucket),
		Prefix:    aws.String(prefix),
		Delimiter: aws.String(delimiter),
	})

	var objects []Object
	var totalSize int64
	var totalCount int

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("list-objects: %w", err)
		}

		// Add common prefixes (directories)
		for _, p := range page.CommonPrefixes {
			objects = append(objects, Object{
				Key: aws.ToString(p.Prefix),
			})
		}

		// Add objects
		for _, obj := range page.Contents {
			objects = append(objects, Object{
				Key:          aws.ToString(obj.Key),
				LastModified: aws.ToTime(obj.LastModified),
				Size:         aws.ToInt64(obj.Size),
				StorageClass: string(obj.StorageClass),
				ETag:         aws.ToString(obj.ETag),
			})
			totalSize += aws.ToInt64(obj.Size)
			totalCount++
		}
	}

	if opts.Summarize {
		return c.printer.PrintJSON(map[string]any{
			"Objects":    objects,
			"TotalSize":  totalSize,
			"TotalCount": totalCount,
		})
	}

	return c.printer.PrintJSON(objects)
}

// CpOptions configures the cp operation
type CpOptions struct {
	Recursive bool
	DryRun    bool
	Quiet     bool
}

// Cp copies files between local and S3
func (c *Client) Cp(ctx context.Context, w io.Writer, src, dst string, opts CpOptions) error {
	srcURI, err := ParseS3URI(src)
	if err != nil {
		return err
	}
	dstURI, err := ParseS3URI(dst)
	if err != nil {
		return err
	}

	switch {
	case srcURI.IsS3 && dstURI.IsS3:
		// S3 to S3 copy
		return c.copyS3ToS3(ctx, w, srcURI, dstURI, opts)
	case srcURI.IsS3:
		// Download from S3
		return c.downloadFromS3(ctx, w, srcURI, dst, opts)
	case dstURI.IsS3:
		// Upload to S3
		return c.uploadToS3(ctx, w, src, dstURI, opts)
	default:
		return fmt.Errorf("at least one argument must be an S3 URI")
	}
}

func (c *Client) copyS3ToS3(ctx context.Context, w io.Writer, src, dst *S3URI, opts CpOptions) error {
	if opts.DryRun {
		_, _ = fmt.Fprintf(w, "(dryrun) copy: s3://%s/%s to s3://%s/%s\n", src.Bucket, src.Key, dst.Bucket, dst.Key)
		return nil
	}

	_, err := c.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(dst.Bucket),
		Key:        aws.String(dst.Key),
		CopySource: aws.String(url.PathEscape(src.Bucket + "/" + src.Key)),
	})
	if err != nil {
		return fmt.Errorf("copy-object: %w", err)
	}

	if !opts.Quiet {
		_, _ = fmt.Fprintf(w, "copy: s3://%s/%s to s3://%s/%s\n", src.Bucket, src.Key, dst.Bucket, dst.Key)
	}
	return nil
}

func (c *Client) downloadFromS3(ctx context.Context, w io.Writer, src *S3URI, dst string, opts CpOptions) error {
	if opts.DryRun {
		_, _ = fmt.Fprintf(w, "(dryrun) download: s3://%s/%s to %s\n", src.Bucket, src.Key, dst)
		return nil
	}

	result, err := c.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(src.Bucket),
		Key:    aws.String(src.Key),
	})
	if err != nil {
		return fmt.Errorf("get-object: %w", err)
	}
	defer func() { _ = result.Body.Close() }()

	// Create destination directory if needed
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}

	file, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer func() { _ = file.Close() }()

	_, err = io.Copy(file, result.Body)
	if err != nil {
		return fmt.Errorf("writing file: %w", err)
	}

	if !opts.Quiet {
		_, _ = fmt.Fprintf(w, "download: s3://%s/%s to %s\n", src.Bucket, src.Key, dst)
	}
	return nil
}

func (c *Client) uploadToS3(ctx context.Context, w io.Writer, src string, dst *S3URI, opts CpOptions) error {
	if opts.DryRun {
		_, _ = fmt.Fprintf(w, "(dryrun) upload: %s to s3://%s/%s\n", src, dst.Bucket, dst.Key)
		return nil
	}

	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("opening file: %w", err)
	}
	defer func() { _ = file.Close() }()

	_, err = c.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(dst.Bucket),
		Key:    aws.String(dst.Key),
		Body:   file,
	})
	if err != nil {
		return fmt.Errorf("put-object: %w", err)
	}

	if !opts.Quiet {
		_, _ = fmt.Fprintf(w, "upload: %s to s3://%s/%s\n", src, dst.Bucket, dst.Key)
	}
	return nil
}

// Rm removes objects from S3
func (c *Client) Rm(ctx context.Context, w io.Writer, uri string, recursive, dryRun, quiet bool) error {
	s3uri, err := ParseS3URI(uri)
	if err != nil {
		return err
	}

	if !s3uri.IsS3 {
		return fmt.Errorf("invalid S3 URI: %s", uri)
	}

	if recursive {
		return c.rmRecursive(ctx, w, s3uri, dryRun, quiet)
	}

	if dryRun {
		_, _ = fmt.Fprintf(w, "(dryrun) delete: s3://%s/%s\n", s3uri.Bucket, s3uri.Key)
		return nil
	}

	_, err = c.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s3uri.Bucket),
		Key:    aws.String(s3uri.Key),
	})
	if err != nil {
		return fmt.Errorf("delete-object: %w", err)
	}

	if !quiet {
		_, _ = fmt.Fprintf(w, "delete: s3://%s/%s\n", s3uri.Bucket, s3uri.Key)
	}
	return nil
}

func (c *Client) rmRecursive(ctx context.Context, w io.Writer, s3uri *S3URI, dryRun, quiet bool) error {
	paginator := s3.NewListObjectsV2Paginator(c.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s3uri.Bucket),
		Prefix: aws.String(s3uri.Key),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("list-objects: %w", err)
		}

		if len(page.Contents) == 0 {
			continue
		}

		// Batch delete
		var objects []types.ObjectIdentifier
		for _, obj := range page.Contents {
			if dryRun {
				_, _ = fmt.Fprintf(w, "(dryrun) delete: s3://%s/%s\n", s3uri.Bucket, aws.ToString(obj.Key))
			} else {
				objects = append(objects, types.ObjectIdentifier{Key: obj.Key})
			}
		}

		if !dryRun && len(objects) > 0 {
			_, err = c.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
				Bucket: aws.String(s3uri.Bucket),
				Delete: &types.Delete{Objects: objects},
			})
			if err != nil {
				return fmt.Errorf("delete-objects: %w", err)
			}

			if !quiet {
				for _, obj := range objects {
					_, _ = fmt.Fprintf(w, "delete: s3://%s/%s\n", s3uri.Bucket, aws.ToString(obj.Key))
				}
			}
		}
	}

	return nil
}

// Mb creates a new bucket
func (c *Client) Mb(ctx context.Context, w io.Writer, uri string, region string) error {
	s3uri, err := ParseS3URI(uri)
	if err != nil {
		return err
	}

	if !s3uri.IsS3 {
		return fmt.Errorf("invalid S3 URI: %s", uri)
	}

	input := &s3.CreateBucketInput{
		Bucket: aws.String(s3uri.Bucket),
	}

	// LocationConstraint is required for regions other than us-east-1
	if region != "" && region != "us-east-1" {
		input.CreateBucketConfiguration = &types.CreateBucketConfiguration{
			LocationConstraint: types.BucketLocationConstraint(region),
		}
	}

	_, err = c.client.CreateBucket(ctx, input)
	if err != nil {
		return fmt.Errorf("create-bucket: %w", err)
	}

	_, _ = fmt.Fprintf(w, "make_bucket: s3://%s\n", s3uri.Bucket)
	return nil
}

// Rb removes a bucket
func (c *Client) Rb(ctx context.Context, w io.Writer, uri string, force bool) error {
	s3uri, err := ParseS3URI(uri)
	if err != nil {
		return err
	}

	if !s3uri.IsS3 {
		return fmt.Errorf("invalid S3 URI: %s", uri)
	}

	if force {
		// Delete all objects first
		if err := c.rmRecursive(ctx, w, s3uri, false, true); err != nil {
			return err
		}
	}

	_, err = c.client.DeleteBucket(ctx, &s3.DeleteBucketInput{
		Bucket: aws.String(s3uri.Bucket),
	})
	if err != nil {
		return fmt.Errorf("delete-bucket: %w", err)
	}

	_, _ = fmt.Fprintf(w, "remove_bucket: s3://%s\n", s3uri.Bucket)
	return nil
}

// PresignOptions configures the presign operation
type PresignOptions struct {
	ExpiresIn time.Duration
}

// Presign generates a presigned URL
func (c *Client) Presign(ctx context.Context, uri string, opts PresignOptions) (string, error) {
	s3uri, err := ParseS3URI(uri)
	if err != nil {
		return "", err
	}

	if !s3uri.IsS3 {
		return "", fmt.Errorf("invalid S3 URI: %s", uri)
	}

	if opts.ExpiresIn == 0 {
		opts.ExpiresIn = 15 * time.Minute
	}

	result, err := c.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s3uri.Bucket),
		Key:    aws.String(s3uri.Key),
	}, func(po *s3.PresignOptions) {
		po.Expires = opts.ExpiresIn
	})
	if err != nil {
		return "", fmt.Errorf("presign: %w", err)
	}

	return result.URL, nil
}
