package pingservices

import (
	"github.com/snappy-fix-golang/external/thirdparty/ipstack"
	"github.com/snappy-fix-golang/pkg/geo"
)

func ReturnTrue() bool {
	return true
}

// ResolveIP tries ipstack → ip-api → ipdata (in that order)
// and returns the first successful result.
func (s Service) ResolveIP(ip string) (ipstack.IPStackResolveIPResponse, error) {
	// choose acceptance policy; set to nil for "first success wins"
	opts := &geo.Options{
		RequireCityOrRegion: true, // set true if you require city/region
		RequireCoordinates:  true, // set true if you require non-zero lat/lon
	}
	return geo.ResolveIPWithFallback(s.ExtReq, ip, opts)
}
