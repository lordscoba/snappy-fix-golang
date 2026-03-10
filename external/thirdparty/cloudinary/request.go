package cloudinary

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/cloudinary/cloudinary-go/v2"
	clduploader "github.com/cloudinary/cloudinary-go/v2/api/uploader"

	logutil "github.com/snappy-fix-golang/pkg/logger"
)

type RequestObj struct {
	Name      string
	CloudName string
	APIKey    string
	APISecret string
	Logger    *logutil.Logger
}

func (r *RequestObj) load() (*cloudinary.Cloudinary, error) {

	cld, err := cloudinary.NewFromParams(
		r.CloudName,
		r.APIKey,
		r.APISecret,
	)

	if err != nil {
		return nil, fmt.Errorf("%s: failed initializing cloudinary: %w", r.Name, err)
	}

	return cld, nil
}

//////////////////////////////////////////////////////
//// Upload Image
//////////////////////////////////////////////////////

func (r *RequestObj) UploadImage(data interface{}) (UploadResponse, error) {

	var resp UploadResponse

	in, ok := data.(UploadInput)

	if !ok {
		return resp, fmt.Errorf("%s: invalid request data (expected UploadInput)", r.Name)
	}

	if len(in.Bytes) == 0 {
		return resp, fmt.Errorf("cloudinary upload: empty payload")
	}

	cld, err := r.load()

	if err != nil {
		return resp, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	reader := bytes.NewReader(in.Bytes)

	params := clduploader.UploadParams{
		Folder:   in.Folder,
		PublicID: in.PublicID,
	}

	result, err := cld.Upload.Upload(ctx, reader, params)

	if err != nil {

		r.Logger.Error("cloudinary upload", "err", err.Error())

		return resp, fmt.Errorf("cloudinary: upload failed: %w", err)
	}

	resp = UploadResponse{
		PublicID:  result.PublicID,
		URL:       result.URL,
		SecureURL: result.SecureURL,
		Width:     result.Width,
		Height:    result.Height,
		Format:    result.Format,
		Resource:  result.ResourceType,
		CreatedAt: result.CreatedAt,
	}

	return resp, nil
}

//////////////////////////////////////////////////////
//// Delete Image
//////////////////////////////////////////////////////

func (r *RequestObj) DeleteImage(data interface{}) (DeleteResponse, error) {

	var resp DeleteResponse

	in, ok := data.(DeleteInput)

	if !ok || in.PublicID == "" {

		return resp, fmt.Errorf("%s: invalid request data (expected DeleteInput)", r.Name)
	}

	cld, err := r.load()

	if err != nil {
		return resp, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := cld.Upload.Destroy(ctx, clduploader.DestroyParams{
		PublicID: in.PublicID,
	})

	if err != nil {

		r.Logger.Error("cloudinary delete", "public_id", in.PublicID, "err", err.Error())

		return resp, fmt.Errorf("cloudinary delete failed: %w", err)
	}

	resp.Result = result.Result

	return resp, nil
}

//////////////////////////////////////////////////////
//// Get Image URL
//////////////////////////////////////////////////////

func (r *RequestObj) GetURL(data interface{}) (GetURLResponse, error) {

	var resp GetURLResponse

	in, ok := data.(GetURLInput)

	if !ok || in.PublicID == "" {

		return resp, fmt.Errorf("%s: invalid request data (expected GetURLInput)", r.Name)
	}

	cld, err := r.load()

	if err != nil {
		return resp, err
	}

	url, err := cld.Image(in.PublicID)

	if err != nil {
		return resp, fmt.Errorf("cloudinary get url failed: %w", err)
	}

	// Transformation:
	// f_auto = automatic format (WebP, AVIF etc)
	// q_auto = automatic quality optimization
	// c_limit = prevents overscaling
	// url.Transformation = "f_auto,q_auto,c_limit,w_1200"

	urlStr, err := url.String()
	if err != nil {
		return resp, fmt.Errorf("cloudinary get url failed: %w", err)
	}

	resp.URL = urlStr

	return resp, nil
}

// Upload image:

// resp, err := externalRequest.SendExternalRequestS3Store(
// 	external.CloudinaryUploadImage,
// 	cloudinary.UploadInput{
// 		Bytes: imageBytes,
// 		Folder: "news",
// 	},
// )

// Delete image::

// externalRequest.SendExternalRequestS3Store(
// 	external.CloudinaryDeleteImage,
// 	cloudinary.DeleteInput{
// 		PublicID: "news/abc123",
// 	},
// )

// Get image URL:

// resp, err := externalRequest.SendExternalRequestS3Store(
// 	external.CloudinaryGetURL,
// 	cloudinary.GetURLInput{
// 		PublicID: "news/abc123",
// 	},
// )
