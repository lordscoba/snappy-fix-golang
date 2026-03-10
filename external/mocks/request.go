package mocks

import (
	"fmt"

	"github.com/snappy-fix-golang/external"
	ipapimocks "github.com/snappy-fix-golang/external/mocks/ipapi_mocks"
	ipdatamocks "github.com/snappy-fix-golang/external/mocks/ipdata_mocks"
	ipstackmocks "github.com/snappy-fix-golang/external/mocks/ipstack_mocks"

	logutil "github.com/snappy-fix-golang/pkg/logger"
)

type ExternalRequest struct {
	Logger     *logutil.Logger
	Test       bool
	RequestObj RequestObj
}

type RequestObj struct {
	Name         string
	Path         string
	Method       string
	Headers      map[string]string
	SuccessCode  int
	RequestData  interface{}
	DecodeMethod string
	Logger       *logutil.Logger
}

var (
	JsonDecodeMethod    = "json"
	PhpSerializerMethod = "phpserializer"
)

func (er ExternalRequest) SendExternalRequest(name string, data interface{}) (interface{}, error) {
	switch name {
	case external.IpstackResolveIp:
		return ipstackmocks.IpstackResolveIp(er.Logger, data)
	case external.IpdataResolveIp:
		return ipdatamocks.IpdataResolveIp(er.Logger, data)
	case external.IpapiResolveIp:
		return ipapimocks.IpapiResolveIp(er.Logger, data)
	default:
		return nil, fmt.Errorf("mocks: request not found: %s", name)
	}
}
