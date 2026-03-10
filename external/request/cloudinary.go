package request

import (
	"fmt"

	"github.com/snappy-fix-golang/external"
	"github.com/snappy-fix-golang/external/mocks"
	"github.com/snappy-fix-golang/external/thirdparty/cloudinary"
	"github.com/snappy-fix-golang/internal/config"
)

func (er ExternalRequest) SendExternalRequestCloudinary(name string, data interface{}) (interface{}, error) {
	cfg := config.GetConfig()

	if !er.Test {
		switch name {
		case external.CloudinaryUploadImage:

			obj := cloudinary.RequestObj{
				Name:      name,
				CloudName: cfg.Cloudinary.CloudName,
				APIKey:    cfg.Cloudinary.APIKey,
				APISecret: cfg.Cloudinary.APISecret,
				Logger:    er.Logger,
			}

			return obj.UploadImage(data)

		case external.CloudinaryDeleteImage:

			obj := cloudinary.RequestObj{
				Name:      name,
				CloudName: cfg.Cloudinary.CloudName,
				APIKey:    cfg.Cloudinary.APIKey,
				APISecret: cfg.Cloudinary.APISecret,
				Logger:    er.Logger,
			}

			return obj.DeleteImage(data)

		case external.CloudinaryGetURL:

			obj := cloudinary.RequestObj{
				Name:      name,
				CloudName: cfg.Cloudinary.CloudName,
				APIKey:    cfg.Cloudinary.APIKey,
				APISecret: cfg.Cloudinary.APISecret,
				Logger:    er.Logger,
			}

			return obj.GetURL(data)

		default:
			return nil, fmt.Errorf("request: unsupported request name: %s", name)
		}
	}

	// Test path: delegate to mocks
	mer := mocks.ExternalRequest{Logger: er.Logger, Test: true}
	return mer.SendExternalRequest(name, data)
}
