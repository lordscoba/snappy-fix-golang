package external

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/elliotchance/phpserialize"
	logutil "github.com/snappy-fix-golang/pkg/logger"
)

var (
	// requests
	IpstackResolveIp = "ipstack_resolve_ip"
	IpdataResolveIp  = "ipdata_resolve_ip"
	IpapiResolveIp   = "ipapi_resolve_ip"

	// mailgun
	MailgunSend = "mailgun_send"

	S3UploadObject = "s3_upload_object"
	S3DeleteObject = "s3_delete_object"
	S3PresignGet   = "s3_presign_get"
	S3GetObject    = "s3_get_object"

	// paystack
	PaystackInit             = "paystack_init"
	PaystackVerify           = "paystack_verify"
	PaystackCreateRecipient  = "paystack_create_recipient"
	PaystackTransfer         = "paystack_transfer"
	PaystackListBanks        = "paystack_list_banks"
	PaystackCreateSubaccount = "paystack_create_subaccount"
	PaystackCreateSplit      = "paystack_create_split"

	CloudinaryUploadImage = "cloudinary_upload_image"
	CloudinaryDeleteImage = "cloudinary_delete_image"
	CloudinaryGetURL      = "cloudinary_get_url"
)

type SendRequestObject struct {
	Name         string
	Logger       *logutil.Logger
	Path         string            // Base URL (scheme://host[:port][basepath])
	Method       string            // GET, POST, etc.
	Headers      map[string]string // Request headers
	SuccessCode  int               // Expected success status (exact match is preferred)
	Data         interface{}       // Optional request payload (JSON-encoded when non-nil)
	DecodeMethod string            // "json" | "phpserializer"
	UrlPrefix    string            // Typically the path+query to append to Path

	// Optional overrides for testability and robustness
	Client  *http.Client    // If nil, a default client with timeout will be used
	Context context.Context // If nil, context.Background() with timeout is used
	Timeout time.Duration   // If zero, a sensible default is used

	// NEW (optional): use when you need a non-JSON body (e.g., Mailgun form).
	ContentTypeOverride string // e.g., "application/x-www-form-urlencoded"
	BodyBytes           []byte // if set, this is sent as the request body (instead of JSON-encoding Data)
}

// Constructor
func GetNewSendRequestObject(
	logger *logutil.Logger,
	name, path, method, urlPrefix, decodeMethod string,
	headers map[string]string,
	successCode int,
	data interface{},
) *SendRequestObject {
	return &SendRequestObject{
		Logger:       logger,
		Name:         name,
		Path:         path,
		Method:       method,
		UrlPrefix:    urlPrefix,
		DecodeMethod: decodeMethod,
		Headers:      headers,
		SuccessCode:  successCode,
		Data:         data,
	}
}

const defaultTimeout = 12 * time.Second

// SendRequest executes the HTTP request and decodes into 'response' when successful.
// Returns (statusCode, rawBody, error).
func (r *SendRequestObject) SendRequest(response interface{}) (int, []byte, error) {
	logger := r.Logger
	name := r.Name

	// Build final URL: Path (base) + UrlPrefix (path+query). Do NOT log secrets elsewhere.
	finalURL := r.Path
	if r.UrlPrefix != "" {
		finalURL += r.UrlPrefix
	}

	// Choose client with timeout
	client := r.Client
	if client == nil {
		client = &http.Client{Timeout: defaultTimeout}
	}
	// Choose context with timeout
	ctx := r.Context
	if ctx == nil {
		t := r.Timeout
		if t <= 0 {
			t = defaultTimeout
		}
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), t)
		defer cancel()
	}

	var bodyReader io.Reader
	// Prefer pre-encoded BodyBytes if provided (used for form or custom encodings)
	if len(r.BodyBytes) > 0 {
		bodyReader = bytes.NewReader(r.BodyBytes)
	} else if r.Method != http.MethodGet && r.Data != nil {
		// default JSON for non-GET
		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(r.Data); err != nil {
			logger.Error("encoding error", name, err.Error())
			return 0, nil, fmt.Errorf("%s: json encode error: %w", name, err)
		}
		bodyReader = buf
	}

	req, err := http.NewRequestWithContext(ctx, r.Method, finalURL, bodyReader)
	if err != nil {

		logger.Error("request creation error", name, err.Error())
		return 0, nil, fmt.Errorf("%s: request creation error: %w", name, err)
	}

	// Default headers
	// Content-Type logic
	if r.ContentTypeOverride != "" {
		req.Header.Set("Content-Type", r.ContentTypeOverride)
	} else if r.Method != http.MethodGet && r.Data != nil && r.DecodeMethod == JsonDecodeMethod && len(r.BodyBytes) == 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	// if r.Method != http.MethodGet && r.Data != nil && r.DecodeMethod == JsonDecodeMethod {
	// 	// Set content type only when sending JSON body
	// 	req.Header.Set("Content-Type", "application/json")
	// }
	for k, v := range r.Headers {
		req.Header.Set(k, v)
	}

	logger.Info("request start", "name", name, "method", r.Method)

	// fmt.Println("req", req)

	res, err := client.Do(req)
	if err != nil {
		logger.Error("client do", name, err.Error())
		return 0, nil, fmt.Errorf("%s: http client error: %w", name, err)
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Error("reading body error", name, err.Error())
		return res.StatusCode, nil, fmt.Errorf("%s: read body error: %w", name, err)
	}

	// fmt.Println("raw", string(raw), "statusCode", res.StatusCode, "error", err)

	// Check status BEFORE decoding
	if res.StatusCode != r.SuccessCode {
		// Treat any non-2xx or unexpected code as error; include a safe snippet of the body.
		snippet := raw
		if len(snippet) > 256 {
			snippet = snippet[:256]
		}
		return res.StatusCode, raw, fmt.Errorf("%s: unexpected status %d; body: %s", name, res.StatusCode, string(snippet))
	}

	// Decode successful response
	switch r.DecodeMethod {
	case JsonDecodeMethod:
		if response != nil {
			if err := json.Unmarshal(raw, response); err != nil {
				logger.Error("json decoding error", name, err.Error())
				return res.StatusCode, raw, fmt.Errorf("%s: json decode error: %w", name, err)
			}
		}
	case PhpSerializerMethod:
		if response != nil {
			if err := phpserialize.Unmarshal(raw, response); err != nil {
				logger.Error("php serializer decoding error", name, err.Error())
				return res.StatusCode, raw, fmt.Errorf("%s: php unserialize error: %w", name, err)
			}
		}
	default:
		// If a decode method is unknown, return the raw body and an error.
		return res.StatusCode, raw, fmt.Errorf("%s: unsupported decode method: %s", name, r.DecodeMethod)
	}

	logger.Info("request success", "name", name, "status", res.StatusCode)
	return res.StatusCode, raw, nil
}

// Decode method constants (kept for compatibility with other packages)
const (
	JsonDecodeMethod    = "json"
	PhpSerializerMethod = "phpserializer"
)
