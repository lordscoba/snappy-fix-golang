package ipdatamocks

import (
	"fmt"

	"github.com/snappy-fix-golang/external/thirdparty/ipstack" // reuse unified struct
	logutil "github.com/snappy-fix-golang/pkg/logger"
)

func IpdataResolveIp(logger *logutil.Logger, idata interface{}) (ipstack.IPStackResolveIPResponse, error) {
	var out ipstack.IPStackResolveIPResponse

	ip, ok := idata.(string)
	if !ok {
		logger.Error("ipdata resolve ip (mock)", idata, "request data format error (expected string IP)")
		return out, fmt.Errorf("ipdata_resolve_ip: request data format error (expected string IP)")
	}

	out.Ip = ip
	out.City = "mock-city"
	out.CountryName = "mock-country"
	out.Latitude = 1.23
	out.Longitude = 4.56

	logger.Info("ipdata resolve ip (mock)", "ip", ip)
	return out, nil
}
