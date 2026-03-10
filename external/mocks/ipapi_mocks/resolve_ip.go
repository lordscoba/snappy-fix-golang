package ipapimocks

import (
	"fmt"

	"github.com/snappy-fix-golang/external/thirdparty/ipstack"
	logutil "github.com/snappy-fix-golang/pkg/logger"
)

func IpapiResolveIp(logger *logutil.Logger, idata interface{}) (ipstack.IPStackResolveIPResponse, error) {
	var out ipstack.IPStackResolveIPResponse

	ip, ok := idata.(string)
	if !ok {
		logger.Error("ipapi resolve ip (mock)", idata, "request data format error (expected string IP)")
		return out, fmt.Errorf("ipapi_resolve_ip: request data format error (expected string IP)")
	}

	out.Ip = ip
	out.City = "mock-city"
	out.CountryName = "mock-country"
	out.Latitude = 6.45
	out.Longitude = 3.39

	logger.Info("ipapi resolve ip (mock)", "ip", ip)
	return out, nil
}
