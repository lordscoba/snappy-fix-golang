package request

import (
	"fmt"

	"github.com/snappy-fix-golang/external"
	"github.com/snappy-fix-golang/external/mocks"
	"github.com/snappy-fix-golang/external/thirdparty/s3store"
	"github.com/snappy-fix-golang/internal/config"
)

func (er ExternalRequest) SendExternalRequestS3Store(name string, data interface{}) (interface{}, error) {
	cfg := config.GetConfig()

	if !er.Test {
		switch name {
		// ---------- S3 operations ----------
		case external.S3UploadObject:
			obj := s3store.RequestObj{
				Name:       name,
				Bucket:     cfg.AmazonS3Bucket.Bucket,     // make sure your config maps AWS_S3_BUCKET
				Region:     cfg.AmazonS3Bucket.Region,     // AWS_REGION
				PublicBase: cfg.AmazonS3Bucket.PublicBase, // optional; to build public URLs
				Logger:     er.Logger,
			}
			return obj.UploadObject(data)

		case external.S3DeleteObject:
			obj := s3store.RequestObj{
				Name:   name,
				Bucket: cfg.AmazonS3Bucket.Bucket,
				Region: cfg.AmazonS3Bucket.Region,
				Logger: er.Logger,
			}
			return obj.DeleteObject(data)

		case external.S3PresignGet:
			obj := s3store.RequestObj{
				Name:   name,
				Bucket: cfg.AmazonS3Bucket.Bucket,
				Region: cfg.AmazonS3Bucket.Region,
				Logger: er.Logger,
			}
			return obj.PresignGet(data)
		case external.S3GetObject:
			obj := s3store.RequestObj{
				Name:       name,
				Bucket:     cfg.AmazonS3Bucket.Bucket,
				Region:     cfg.AmazonS3Bucket.Region,
				PublicBase: cfg.AmazonS3Bucket.PublicBase, // optional
				Logger:     er.Logger,
			}
			return obj.GetObject(data)
		default:
			return nil, fmt.Errorf("request: unsupported request name: %s", name)
		}
	}

	// Test path: delegate to mocks
	mer := mocks.ExternalRequest{Logger: er.Logger, Test: true}
	return mer.SendExternalRequest(name, data)
}
