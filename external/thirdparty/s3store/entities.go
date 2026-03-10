// external/thirdparty/s3store/types.go
package s3store

import (
	"time"
)

type UploadInput struct {
	// Bytes is the binary image payload to upload
	Bytes []byte

	// Suggested filename (used only to derive extension if ContentType is empty)
	Filename string

	// ContentType must be an image content-type (e.g., image/png, image/jpeg)
	ContentType string

	// Prefix (folder) inside bucket (e.g., "avatars/"). Optional.
	Prefix string

	// Optional custom key; if empty, we’ll generate a UUID-based key with inferred extension.
	Key string

	// If true, we’ll generate a presigned GET URL (for private buckets)
	Presign bool

	// TTL for the presigned URL (default 5m)
	PresignTTL time.Duration

	IsImage bool
}

type UploadResponse struct {
	Key         string
	URL         string // public URL (if bucket public) or presigned URL (if requested)
	ETag        string
	Size        int64
	ContentType string
}

type DeleteInput struct {
	Key string
}

type DeleteResponse struct {
	Deleted bool
}

type PresignGetInput struct {
	Key string
	TTL time.Duration
}

type PresignGetResponse struct {
	URL string
}

type GetObjectInput struct {
	Key string `json:"key"`
}

type GetObjectResponse struct {
	Key         string `json:"key"`
	Bytes       []byte `json:"-"`
	ETag        string `json:"etag"`
	Size        int64  `json:"size"`
	ContentType string `json:"content_type"`
	// Optional: a short-lived URL if you want to return one with Get
	PresignedURL string    `json:"presigned_url,omitempty"`
	ExpiresAt    time.Time `json:"expires_at,omitempty"`
}
