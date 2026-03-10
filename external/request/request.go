package request

import (
	"fmt"

	"github.com/snappy-fix-golang/external/mocks"
	"github.com/snappy-fix-golang/internal/config"
	logutil "github.com/snappy-fix-golang/pkg/logger"
)

type ExternalRequest struct {
	Logger *logutil.Logger
	Test   bool
}

var (
	// serializers
	JsonDecodeMethod    = "json"
	PhpSerializerMethod = "phpserializer"
)

func (er ExternalRequest) SendExternalRequestK(name string, data interface{}) (interface{}, error) {
	_ = config.GetConfig()

	if !er.Test {
		switch name {

		default:
			return nil, fmt.Errorf("request: unsupported request name: %s", name)
		}
	}

	// Test path: delegate to mocks
	mer := mocks.ExternalRequest{Logger: er.Logger, Test: true}
	return mer.SendExternalRequest(name, data)
}
