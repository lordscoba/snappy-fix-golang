package ipstack

import (
	"github.com/snappy-fix-golang/external"
	logutil "github.com/snappy-fix-golang/pkg/logger"
)

type RequestObj struct {
	Name         string
	Path         string
	Method       string
	SuccessCode  int
	RequestData  interface{}
	DecodeMethod string
	Logger       *logutil.Logger
}

var (
	JsonDecodeMethod    = "json"
	PhpSerializerMethod = "phpserializer"
)

// getNewSendRequestObject prepares a request sender. A custom HTTP client can be
// injected later via the returned object if needed for tests.
func (r *RequestObj) getNewSendRequestObject(data interface{}, headers map[string]string, urlprefix string) *external.SendRequestObject {
	return external.GetNewSendRequestObject(r.Logger, r.Name, r.Path, r.Method, urlprefix, r.DecodeMethod, headers, r.SuccessCode, data)
}
