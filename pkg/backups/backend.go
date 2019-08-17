package backups

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Backend provides read and write capabilities to a filesystem-like storage.
type Backend interface {
	Store(name string, reader io.Reader) error
	Read(name string) (io.Reader, error)
}

// NoSuchName is an error returned by `Backend` when a name does not exist.
type NoSuchName struct {
	Name string
}

func (e NoSuchName) Error() string {
	return fmt.Sprintf("%s does not exist", e.Name)
}

// S3Backend provides a Backend to AWS S3. This will also work with Digital
// Ocean Spaces, since this product is also S3-compatible.
type S3Backend struct {
	session *session.Session

	// The bucket in which to manage backups.
	bucket string
}

// NewS3Backend returns an S3Backend with the given session and bucket.
func NewS3Backend(session *session.Session, bucket string) S3Backend {
	return S3Backend{
		session: session,
		bucket:  bucket,
	}
}

// Store stores `reader`'s bytes under `name` in S3 under the configured bucket.
func (b S3Backend) Store(name string, reader io.Reader) error {
	svc := s3.New(b.session)

	contents, err := ioutil.ReadAll(reader)
	if err != nil {
		return err
	}

	input := &s3.PutObjectInput{
		Body:   bytes.NewReader(contents),
		Bucket: aws.String(b.bucket),
		Key:    aws.String(name),
	}

	_, err = svc.PutObject(input)

	return err
}

// Read attempts to download `name` from S3 and return a reader.
func (b S3Backend) Read(name string) (io.Reader, error) {
	svc := s3.New(b.session)

	input := &s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(name),
	}

	output, err := svc.GetObject(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case s3.ErrCodeNoSuchKey:
				return nil, NoSuchName{name}
			default:
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return output.Body, nil
}

// InMemoryBackend stores backups in a slice. This should only be used for testing.
type InMemoryBackend struct {
	Backups map[string][]byte
}

func NewInMemoryBackend() *InMemoryBackend {
	return &InMemoryBackend{
		Backups: make(map[string][]byte),
	}
}

func (b *InMemoryBackend) Store(name string, reader io.Reader) error {
	contents, _ := ioutil.ReadAll(reader)
	b.Backups[name] = contents

	return nil
}

func (b *InMemoryBackend) Read(name string) (io.Reader, error) {
	if value, ok := b.Backups[name]; ok {
		reader := bytes.NewReader(value)

		return reader, nil
	}

	return nil, NoSuchName{name}
}
