package ipstackmocks

import (
	"fmt"

	"github.com/snappy-fix-golang/external/thirdparty/ipstack"
	logutil "github.com/snappy-fix-golang/pkg/logger"
)

func IpstackResolveIp(logger *logutil.Logger, idata interface{}) (ipstack.IPStackResolveIPResponse, error) {
	var out ipstack.IPStackResolveIPResponse

	ip, ok := idata.(string)
	if !ok {
		logger.Error("ipstack resolve ip", idata, "request data format error: expected string IP")
		return out, fmt.Errorf("ipstack_resolve_ip: request data format error (expected string IP)")
	}

	// Provide stable, deterministic mock values.
	out.Ip = ip
	out.City = "city"
	out.CountryName = "name"

	// Avoid logging secrets/URLs in mocks as well.
	logger.Info("ipstack resolve ip (mock)", "ip", ip)

	return out, nil
}
