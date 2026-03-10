// pkg/storage/s3images/service.go
package s3images

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/snappy-fix-golang/external"
	"github.com/snappy-fix-golang/external/request"
	"github.com/snappy-fix-golang/external/thirdparty/s3store"
)

type Service struct {
	ExtReq request.ExternalRequest
}

// CreateImage uploads a new image and returns key + URL (+ metadata)
func (s Service) CreateImage(ctx context.Context, in s3store.UploadInput) (s3store.UploadResponse, error) {
	// basic validation
	if len(in.Bytes) == 0 {
		return s3store.UploadResponse{}, errors.New("empty image")
	}
	// we don’t pass ctx through ExternalRequest today; if you need strict deadlines,
	// add WithContext to ext transport or pass TTL in UploadInput.PresignTTL.
	respAny, err := s.ExtReq.SendExternalRequestS3Store(external.S3UploadObject, in)
	if err != nil {
		return s3store.UploadResponse{}, fmt.Errorf("s3: create image failed: %w", err)
	}
	out, ok := respAny.(s3store.UploadResponse)
	if !ok {
		return s3store.UploadResponse{}, fmt.Errorf("s3: unexpected upload response type")
	}
	return out, nil
}

// UpdateImage deletes the old key (best-effort), then uploads the new bytes and returns its info.
// If deletion fails, returns an error and DOES NOT upload (safer default). Change if you prefer.
func (s Service) UpdateImage(ctx context.Context, oldKey string, in s3store.UploadInput) (s3store.UploadResponse, error) {
	// 1) delete old
	if oldKey != "" {
		_, err := s.ExtReq.SendExternalRequestS3Store(external.S3DeleteObject, s3store.DeleteInput{Key: oldKey})
		if err != nil {
			return s3store.UploadResponse{}, fmt.Errorf("s3: delete old image failed: %w", err)
		}
	}

	// 2) upload new
	return s.CreateImage(ctx, in)
}

// DeleteImage removes the object by key
func (s Service) DeleteImage(ctx context.Context, key string) error {
	_, err := s.ExtReq.SendExternalRequestS3Store(external.S3DeleteObject, s3store.DeleteInput{Key: key})
	return err
}

// Optional convenience: Get one-time presigned URL for a private object
func (s Service) Presign(ctx context.Context, key string, ttl time.Duration) (string, error) {
	respAny, err := s.ExtReq.SendExternalRequestS3Store(external.S3PresignGet, s3store.PresignGetInput{
		Key: key, TTL: ttl,
	})
	if err != nil {
		return "", err
	}
	out, ok := respAny.(s3store.PresignGetResponse)
	if !ok {
		return "", fmt.Errorf("s3: unexpected presign response type")
	}
	return out.URL, nil
}

// GetImage downloads the object bytes + basic metadata.
// If you only need a link, prefer Presign (faster/cheaper).
func (s Service) GetImage(ctx context.Context, key string) (s3store.GetObjectResponse, error) {
	if key == "" {
		return s3store.GetObjectResponse{}, errors.New("key is required")
	}
	respAny, err := s.ExtReq.SendExternalRequestS3Store(external.S3GetObject, s3store.GetObjectInput{Key: key})
	if err != nil {
		return s3store.GetObjectResponse{}, fmt.Errorf("s3: get image failed: %w", err)
	}
	out, ok := respAny.(s3store.GetObjectResponse)
	if !ok {
		return s3store.GetObjectResponse{}, fmt.Errorf("s3: unexpected get-object response type")
	}
	return out, nil
}
