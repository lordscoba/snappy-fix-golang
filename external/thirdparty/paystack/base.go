package paystack

import (
	"github.com/snappy-fix-golang/external"
	logutil "github.com/snappy-fix-golang/pkg/logger"
)

type RequestObj struct {
	Name         string
	Path         string // Base URL, e.g. https://api.paystack.co
	Method       string
	SuccessCode  int
	RequestData  interface{}
	DecodeMethod string
	Logger       *logutil.Logger
}

func (r *RequestObj) sender(data interface{}, headers map[string]string, urlprefix string) *external.SendRequestObject {
	return external.GetNewSendRequestObject(
		r.Logger, r.Name, r.Path, r.Method, urlprefix, r.DecodeMethod, headers, r.SuccessCode, data,
	)
}
