package s3store

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"path"
	"path/filepath"
	"strings"
	"time"

	awsCfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"

	logutil "github.com/snappy-fix-golang/pkg/logger"
)

type RequestObj struct {
	Name       string
	Bucket     string
	Region     string
	PublicBase string // optional: e.g. "https://nactive-app-images.s3.amazonaws.com"
	Logger     *logutil.Logger
}

// ---------- helpers ----------

func (r *RequestObj) load(cfgRegion string) (*s3.Client, error) {
	cfg, err := awsCfg.LoadDefaultConfig(context.Background(),
		awsCfg.WithRegion(cfgRegion),
	)
	if err != nil {
		return nil, fmt.Errorf("%s: failed loading AWS config: %w", r.Name, err)
	}
	return s3.NewFromConfig(cfg), nil
}

func randHex(n int) string {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func ensureImageContentType(ct, filename string) (string, error) {
	ct = strings.TrimSpace(ct)
	if ct == "" {
		// try to infer from filename
		ext := strings.ToLower(filepath.Ext(filename))
		if ext != "" {
			ct = mime.TypeByExtension(ext)
		}
	}
	if ct == "" {
		return "", errors.New("content-type required or inferable from filename")
	}
	if !strings.HasPrefix(ct, "image/") {
		return "", fmt.Errorf("content-type must be image/*, got %q", ct)
	}
	return ct, nil
}

func ensureFileContentType(ct, filename string) (string, error) {
	ct = strings.TrimSpace(ct)

	if ct == "" {
		// try to infer from filename
		ext := strings.ToLower(filepath.Ext(filename))
		if ext != "" {
			ct = mime.TypeByExtension(ext)
		}
	}
	if ct == "" {
		return "", errors.New("content-type required or inferable from filename")
	}

	// Whitelist of safe content-types
	allowed := map[string]bool{
		// Images
		"image/png":  true,
		"image/jpeg": true,
		"image/gif":  true,
		"image/webp": true,

		// PDF
		"application/pdf": true,

		// Office OpenXML
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   true, // .docx
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true, // .xlsx
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": true, // .pptx

		// Plain text formats
		"text/csv":        true,
		"text/plain":      true,
		"application/rtf": true,
		"text/markdown":   true,
	}

	if _, ok := allowed[ct]; ok {
		return ct, nil
	}

	return "", fmt.Errorf("unsupported content-type %q", ct)
}

// Max image size = 10 MB
const MaxImageSizeBytes = 10 * 1024 * 1024 // 5MB

// ensureImageSize checks that an image payload does not exceed the max allowed size.
func EnsureImageSize(sizeBytes int, maxBytes int) error {
	if sizeBytes <= 0 {
		return errors.New("upload: empty payload")
	}
	if sizeBytes > maxBytes {
		return fmt.Errorf("image too large: %.2f MB > %.2f MB",
			float64(sizeBytes)/(1024*1024),
			float64(maxBytes)/(1024*1024),
		)
	}
	return nil
}

// buildKey generates a storage key for the uploaded file
func buildKey(prefix, key, filename, contentType string) string {
	if key != "" {
		return path.Clean(strings.TrimLeft(key, "/"))
	}

	ext := ""
	if filename != "" {
		ext = strings.ToLower(filepath.Ext(filename))
	}

	if ext == "" {
		// fallback from content-type
		switch contentType {
		case "image/png":
			ext = ".png"
		case "image/jpeg":
			ext = ".jpg"
		case "image/gif":
			ext = ".gif"
		case "image/webp":
			ext = ".webp"
		case "application/pdf":
			ext = ".pdf"
		case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
			ext = ".docx"
		case "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":
			ext = ".xlsx"
		case "application/vnd.openxmlformats-officedocument.presentationml.presentation":
			ext = ".pptx"
		case "text/csv":
			ext = ".csv"
		case "text/plain":
			ext = ".txt"
		case "application/rtf":
			ext = ".rtf"
		case "text/markdown":
			ext = ".md"
		default:
			ext = ".bin"
		}
	}

	id := fmt.Sprintf("%s-%s%s", time.Now().UTC().Format("20060102T150405Z"), randHex(6), ext)

	if prefix != "" {
		return path.Clean(path.Join(prefix, id))
	}
	return id
}

func (r *RequestObj) makeURL(key string, presigned string) string {
	if presigned != "" {
		return presigned
	}
	// If you configured PublicBase, use it; else return s3:// style URL as a fallback
	if strings.TrimSpace(r.PublicBase) != "" {
		return strings.TrimRight(r.PublicBase, "/") + "/" + key
	}
	return "s3://" + r.Bucket + "/" + key
}

// ---------- operations used by the coordinator ----------

// UploadObject expects `data` to be UploadInput
func (r *RequestObj) UploadObject(data interface{}) (UploadResponse, error) {
	var resp UploadResponse
	in, ok := data.(UploadInput)
	if !ok {
		return resp, fmt.Errorf("%s: invalid request data (expected UploadInput)", r.Name)
	}
	if len(in.Bytes) == 0 {
		return resp, errors.New("upload: empty payload")
	}

	// Choose the right content-type validator
	var (
		ct  string
		err error
	)
	if in.IsImage {
		// 1) Validate content-type is image/*
		ct, err = ensureImageContentType(in.ContentType, in.Filename)
		if err != nil {
			return resp, err
		}
		// 2) Enforce max image size (5 MB)
		if err := EnsureImageSize(len(in.Bytes), MaxImageSizeBytes); err != nil {
			return resp, err
		}

	} else {
		ct, err = ensureFileContentType(in.ContentType, in.Filename)
		if err != nil {
			return resp, err
		}

		// 2) Enforce max image size (5 MB)
		if err := EnsureImageSize(len(in.Bytes), MaxImageSizeBytes); err != nil {
			return resp, err
		}
		// Optional: if this branch is meant for *documents only* (no images),
		// uncomment this guard to reject accidental image uploads here.
		// if strings.HasPrefix(ct, "image/") {
		// 	return resp, fmt.Errorf("expected a document upload, got image content-type %q", ct)
		// }
	}
	if err != nil {
		return resp, err
	}

	client, err := r.load(r.Region)
	if err != nil {
		return resp, err
	}

	key := buildKey(in.Prefix, in.Key, in.Filename, ct)

	uploader := manager.NewUploader(client)
	out, err := uploader.Upload(context.Background(), &s3.PutObjectInput{
		Bucket:      &r.Bucket,
		Key:         &key,
		Body:        bytes.NewReader(in.Bytes),
		ContentType: &ct,
		ACL:         s3types.ObjectCannedACLPrivate,
	})
	if err != nil {
		r.Logger.Error("s3 upload", "bucket", r.Bucket, "key", key, "err", err.Error())
		return resp, fmt.Errorf("s3: upload failed: %w", err)
	}

	head, herr := client.HeadObject(context.Background(), &s3.HeadObjectInput{
		Bucket: &r.Bucket,
		Key:    &key,
	})
	if herr != nil {
		r.Logger.Error("s3 head after upload", "bucket", r.Bucket, "key", key, "err", herr.Error())
	}

	var presignedURL string
	if in.Presign {
		ps := s3.NewPresignClient(client)
		ttl := in.PresignTTL
		if ttl <= 0 {
			ttl = 5 * time.Minute
		}
		pre, perr := ps.PresignGetObject(context.Background(), &s3.GetObjectInput{
			Bucket: &r.Bucket,
			Key:    &key,
		}, s3.WithPresignExpires(ttl))
		if perr == nil {
			presignedURL = pre.URL
		} else {
			r.Logger.Error("s3 presign", "bucket", r.Bucket, "key", key, "err", perr.Error())
		}
	}

	resp = UploadResponse{
		Key:         key,
		URL:         r.makeURL(key, presignedURL),
		ETag:        deref(head.ETag),
		Size:        deref64(head.ContentLength),
		ContentType: ct,
	}
	_ = out
	return resp, nil
}

// DeleteObject expects `data` to be DeleteInput
func (r *RequestObj) DeleteObject(data interface{}) (DeleteResponse, error) {
	var resp DeleteResponse
	in, ok := data.(DeleteInput)
	if !ok || strings.TrimSpace(in.Key) == "" {
		return resp, fmt.Errorf("%s: invalid request data (expected s3store.DeleteInput)", r.Name)
	}
	client, err := r.load(r.Region)
	if err != nil {
		return resp, err
	}
	_, err = client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: &r.Bucket,
		Key:    &in.Key,
	})
	if err != nil {
		return resp, fmt.Errorf("s3: delete failed: %w", err)
	}
	resp.Deleted = true
	return resp, nil
}

// PresignGet expects `data` to be PresignGetInput
func (r *RequestObj) PresignGet(data interface{}) (PresignGetResponse, error) {
	var resp PresignGetResponse
	in, ok := data.(PresignGetInput)
	if !ok || strings.TrimSpace(in.Key) == "" {
		return resp, fmt.Errorf("%s: invalid request data (expected s3store.PresignGetInput)", r.Name)
	}
	client, err := r.load(r.Region)
	if err != nil {
		return resp, err
	}
	ttl := in.TTL
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}
	ps := s3.NewPresignClient(client) // ✅ new location
	pre, err := ps.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &r.Bucket,
		Key:    &in.Key,
	}, s3.WithPresignExpires(ttl)) // ✅ new option name
	if err != nil {
		return resp, fmt.Errorf("s3: presign failed: %w", err)
	}
	resp.URL = pre.URL
	return resp, nil
}

func (r *RequestObj) GetObject(data interface{}) (GetObjectResponse, error) {
	var resp GetObjectResponse
	in, ok := data.(GetObjectInput)
	if !ok || in.Key == "" {
		return resp, fmt.Errorf("%s: invalid request data (expected s3store.GetObjectInput)", r.Name)
	}

	client, err := r.load(r.Region)
	if err != nil {
		return resp, err
	}

	out, err := client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: &r.Bucket,
		Key:    &in.Key,
	})
	if err != nil {
		r.Logger.Error("s3 get", "bucket", r.Bucket, "key", in.Key, "err", err.Error())
		return resp, fmt.Errorf("s3: get failed: %w", err)
	}
	defer out.Body.Close()

	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, out.Body); err != nil {
		return resp, fmt.Errorf("s3: read body failed: %w", err)
	}

	resp = GetObjectResponse{
		Key:         in.Key,
		Bytes:       buf.Bytes(),
		ETag:        deref(out.ETag),
		Size:        deref64(out.ContentLength),
		ContentType: deref(out.ContentType),
	}

	// (Optional) also hand back a short-lived URL with the bytes, if you like:
	// ps := s3.NewPresignClient(client)
	// pre, perr := ps.PresignGetObject(context.Background(),
	// 	&s3.GetObjectInput{Bucket: &r.Bucket, Key: &in.Key},
	// 	s3.WithPresignExpires(2*time.Minute),
	// )
	// if perr == nil {
	// 	resp.PresignedURL = pre.URL
	// 	resp.ExpiresAt = time.Now().Add(2 * time.Minute)
	// }

	_ = time.Now // silence import if you comment out the optional block
	return resp, nil
}

func deref(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
func deref64(p *int64) int64 {
	if p == nil {
		return 0
	}
	return *p
}

// func buildKey(prefix, key, filename, contentType string) string {
// 	if key != "" {
// 		return path.Clean(strings.TrimLeft(key, "/"))
// 	}
// 	ext := ""
// 	if filename != "" {
// 		ext = strings.ToLower(filepath.Ext(filename))
// 	}
// 	if ext == "" {
// 		// derive a common extension from content-type
// 		switch contentType {
// 		case "image/png":
// 			ext = ".png"
// 		case "image/jpeg":
// 			ext = ".jpg"
// 		case "image/gif":
// 			ext = ".gif"
// 		case "image/webp":
// 			ext = ".webp"
// 		default:
// 			ext = ".img"
// 		}
// 	}
// 	id := fmt.Sprintf("%s-%s%s", time.Now().UTC().Format("20060102T150405Z"), randHex(6), ext)
// 	if prefix != "" {
// 		return path.Clean(path.Join(prefix, id))
// 	}
// 	return id
// }
