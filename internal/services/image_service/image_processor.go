package imageservice

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"net/http"

	"github.com/disintegration/imaging"
	"github.com/kolesa-team/go-webp/encoder"
	"github.com/kolesa-team/go-webp/webp"
	"github.com/snappy-fix-golang/external/request"
	"github.com/snappy-fix-golang/external/thirdparty/cloudinary"
)

// Define a custom error type for validation/security issues
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

const MaxFileSize = 10 << 20 // 10MB

func ValidateAndOptimize(fileBytes []byte) ([]byte, string, error) {
	// 1. Check Size
	if len(fileBytes) > MaxFileSize {
		return nil, "", &ValidationError{Message: "file size exceeds 5MB limit"}
	}

	// 2. Validate MIME Type (Security)
	contentType := http.DetectContentType(fileBytes)
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	if !validTypes[contentType] {
		return nil, "", &ValidationError{Message: fmt.Sprintf("invalid format: %s", contentType)}
	}

	// 3. Decode & Verify Integrity
	img, _, err := image.Decode(bytes.NewReader(fileBytes))
	if err != nil {
		return nil, "", &ValidationError{Message: "corrupt image data or invalid image content"}
	}

	// 4. Optimization
	img = imaging.Fit(img, 1920, 1080, imaging.Lanczos)

	// 5. Convert to WebP
	var buf bytes.Buffer
	options, err := encoder.NewLossyEncoderOptions(encoder.PresetDefault, 75)
	if err != nil {
		return nil, "", fmt.Errorf("webp encoder setup error: %w", err)
	}

	if err := webp.Encode(&buf, img, options); err != nil {
		return nil, "", fmt.Errorf("webp encoding failed: %w", err)
	}

	return buf.Bytes(), "image/webp", nil
}

func uploadToCloudinary(extReq request.ExternalRequest, b []byte, folder string) (cloudinary.UploadResponse, error) {
	res, err := extReq.SendExternalRequestCloudinary("cloudinary_upload_image", cloudinary.UploadInput{
		Bytes: b, Folder: folder,
	})
	if r, ok := res.(cloudinary.UploadResponse); ok {
		return r, nil
	}
	return cloudinary.UploadResponse{}, err
}

func deleteFromCloudinary(extReq request.ExternalRequest, publicID string) error {
	_, err := extReq.SendExternalRequestCloudinary("cloudinary_delete_image", cloudinary.DeleteInput{PublicID: publicID})
	return err
}
