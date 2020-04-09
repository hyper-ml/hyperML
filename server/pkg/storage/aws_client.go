package storage

import (
	"io"
  "fmt"
  
  "github.com/aws/aws-sdk-go/aws"
  "github.com/aws/aws-sdk-go/aws/credentials"
  "github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/aws/aws-sdk-go/aws/session"

  config_pkg "github.com/hyper-ml/hyperml/server/pkg/config"
	"github.com/hyper-ml/hyperml/server/pkg/base"
   
)


type s3Client struct {
  bucket string
  s3handle *s3.S3
  uploadmgr *s3manager.Uploader
}

func newS3Client(objConfig *config_pkg.ObjStorageConfig) (Client, error){
	cfg := objConfig.S3
	aws_config := aws.Config {
		Region: aws.String(cfg.Region),
	}

	switch {
	case cfg.SecretKey != "":
		aws_config.Credentials = credentials.NewStaticCredentials(cfg.AccessKey, 
			cfg.SecretKey, 
			cfg.SessionToken)
		
		if aws_config.Credentials != nil {
			base.Info("[storage.newS3Client] Aws S3 access granted")
		}

	case cfg.CredPath != "":
		return nil, fmt.Errorf("feature_not_implemented: newS3Client()")

	default: 
		return nil, fmt.Errorf("S3 Credentials not provided in config")
	}
  
	s_options:= session.Options{ Config: aws_config}
	s := session.Must(session.NewSessionWithOptions(s_options))
  
  return &s3Client {
  		bucket: cfg.Bucket,
  		s3handle: s3.New(s),
  		uploadmgr: s3manager.NewUploader(s),
  }, nil 
}

func (s *s3Client) Writer(path string) (io.WriteCloser, error){
	return newAwsWriter(s, path)
}

type awsWriter struct {
	failed chan error
	writer *io.PipeWriter
}

func newAwsWriter(s *s3Client, objpath string) (*awsWriter, error) {
	pipe_reader, pipe_writer := io.Pipe()
	w := &awsWriter {
		failed : make(chan error),
		writer: pipe_writer,
	}

	// start a s3 session that readers from pipe_reader
	go func() {
		_, err := s.uploadmgr.Upload(&s3manager.UploadInput{
			Bucket: aws.String(s.bucket),
			Body: pipe_reader,
			Key: aws.String(objpath),
			ContentEncoding: aws.String("application/octet-stream"),
		})

		// redirect errors to writer channel 
		w.failed <- err
	}()

	return w, nil
}


func (w *awsWriter) Write(p []byte) (int, error) {
	return w.writer.Write(p)
}

func (w *awsWriter) Close() error {
	if err := w.writer.Close(); err!= nil {
		return err
	}
	return <- w.failed
}

func (s *s3Client) Reader(name string, offset int64, size int64) (io.ReadCloser, error) {
	result, err := s.s3handle.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key: aws.String(name),
	})

	if err != nil {
		return nil, err
	}

	return result.Body, nil
}

func (s *s3Client) Delete(name string) error {
	_, err := s.s3handle.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key: aws.String(name),
		})
	return err
}

func (s *s3Client) Exists(name string) bool {
	
	_, err := s.s3handle.HeadObject(&s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(name),
	})

	if err != nil {
		return false
	}

	return true
}

func (c *s3Client) Size(name string) (int64, error) {
	return 0, fmt.Errorf("feature_not_implemented: s3Client.Size()")
}

func (s *s3Client) IsNotExist(err error) (result bool) {
	base.Error("feature_not_implemented: s3Client.IsNotExist()")

	return false
}

func (s *s3Client) SignedURL(op, objname string) (string, error) {
	return "", fmt.Errorf("feature_not_implemented")
}

func (s *s3Client) Merge(parentDir, dest string, src []string) error{
	return fmt.Errorf("feature_not_implemented")
}

