package request

import (
	"fmt"

	"github.com/snappy-fix-golang/external"
	"github.com/snappy-fix-golang/external/mocks"
	"github.com/snappy-fix-golang/external/thirdparty/ipapi"
	"github.com/snappy-fix-golang/external/thirdparty/ipdata"
	"github.com/snappy-fix-golang/external/thirdparty/ipstack"
	"github.com/snappy-fix-golang/internal/config"
)

func (er ExternalRequest) SendExternalRequestIPStack(name string, data interface{}) (interface{}, error) {
	cfg := config.GetConfig()

	if !er.Test {
		switch name {
		case external.IpstackResolveIp:
			obj := ipstack.RequestObj{
				Name:         name,
				Path:         fmt.Sprintf("%v", cfg.IPStack.BaseUrl),
				Method:       "GET",
				SuccessCode:  200,
				DecodeMethod: JsonDecodeMethod,
				RequestData:  data,
				Logger:       er.Logger,
			}
			return obj.IpstackResolveIp()
		case external.IpdataResolveIp:
			obj := ipdata.RequestObj{
				Name:         name,
				Path:         fmt.Sprintf("%v", cfg.IPData.BaseUrl),
				Method:       "GET",
				SuccessCode:  200,
				DecodeMethod: JsonDecodeMethod,
				RequestData:  data,
				Logger:       er.Logger,
			}
			return obj.IpdataResolveIp()
		case external.IpapiResolveIp:
			obj := ipapi.RequestObj{
				Name:         name,
				Path:         fmt.Sprintf("%v", cfg.IPAPI.BaseUrl),
				Method:       "GET",
				SuccessCode:  200, // ip-api returns 200 even on fail; we check raw.Status inside resolver
				DecodeMethod: JsonDecodeMethod,
				RequestData:  data,
				Logger:       er.Logger,
			}
			return obj.IpapiResolveIp()
		default:
			return nil, fmt.Errorf("request: unsupported request name: %s", name)
		}
	}

	// Test path: delegate to mocks
	mer := mocks.ExternalRequest{Logger: er.Logger, Test: true}
	return mer.SendExternalRequest(name, data)
}
