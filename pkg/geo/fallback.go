package geo

import (
	"errors"
	"fmt"
	"strings"

	"github.com/snappy-fix-golang/external"
	"github.com/snappy-fix-golang/external/request"
	"github.com/snappy-fix-golang/external/thirdparty/ipstack"
)

// Options controls acceptance criteria for a "successful" provider result.
type Options struct {
	// If true, the result must have non-empty City or RegionName.
	// Useful when you want more granular accuracy (many mobile IPs won't have these).
	RequireCityOrRegion bool
	// If true, the result must have non-zero coordinates.
	RequireCoordinates bool
}

// ResolveIPWithFallback tries providers in priority order:
// 1) ipstack, 2) ip-api, 3) ipdata.
// It returns the first successful typed result that passes the optional Options checks.
func ResolveIPWithFallback(extReq request.ExternalRequest, ip string, opts *Options) (ipstack.IPStackResolveIPResponse, error) {
	var out ipstack.IPStackResolveIPResponse

	// Providers in priority order.
	providers := []string{
		external.IpdataResolveIp,
		external.IpapiResolveIp,
		external.IpstackResolveIp,
	}

	var errs []string

	for _, name := range providers {
		resp, err := extReq.SendExternalRequestIPStack(name, ip)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", name, err))
			continue
		}

		typed, ok := resp.(ipstack.IPStackResolveIPResponse)
		if !ok {
			errs = append(errs, fmt.Sprintf("%s: unexpected response type", name))
			continue
		}

		// If you want stricter acceptance, enforce here:
		if opts != nil {
			if opts.RequireCityOrRegion && (strings.TrimSpace(typed.City) == "" && strings.TrimSpace(typed.RegionName) == "") {
				errs = append(errs, fmt.Sprintf("%s: missing city/region", name))
				continue
			}
			if opts.RequireCoordinates && (typed.Latitude == 0 && typed.Longitude == 0) {
				errs = append(errs, fmt.Sprintf("%s: missing coordinates", name))
				continue
			}
		}

		// Accepted
		return typed, nil
	}

	if len(errs) == 0 {
		return out, errors.New("no providers configured")
	}
	return out, fmt.Errorf("all providers failed: %s", strings.Join(errs, " | "))
}
