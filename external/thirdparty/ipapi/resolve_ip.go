package ipapi

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/snappy-fix-golang/external/thirdparty/ipstack" // [REUSABLE-3]
	"github.com/snappy-fix-golang/internal/config"
)

// IpapiResolveIp calls ip-api.com and maps its payload to the unified IPStackResolveIPResponse.
func (r *RequestObj) IpapiResolveIp() (ipstack.IPStackResolveIPResponse, error) {
	var (
		_    = config.GetConfig()
		resp ipstack.IPStackResolveIPResponse
		log  = r.Logger
		in   = r.RequestData
	)

	ip, ok := in.(string)
	if !ok {
		log.Error("ipapi resolve ip", in, "request data format error: expected string IP")
		return resp, fmt.Errorf("ipapi_resolve_ip: request data format error (expected string IP)")
	}

	// Build URL safely: {BaseUrl}/json/{ip}?fields=...
	// Note: free plan is HTTP; if you use Pro, set HTTPS in BaseUrl.
	u, err := url.Parse(r.Path) // e.g., http://ip-api.com
	if err != nil {
		log.Error("ipapi resolve ip", "invalid base url", r.Path, err.Error())
		return resp, fmt.Errorf("ipapi_resolve_ip: invalid base URL")
	}
	u.Path = path.Join(u.Path, "json", ip)

	// Reduce payload to only needed fields. ip-api supports CSV field list.
	// We also ask for status & message to handle provider-level failures.
	fields := "status,message,query,continent,continentCode,country,countryCode,region,regionName,city,zip,lat,lon"
	q := u.Query()
	q.Set("fields", fields)
	u.RawQuery = q.Encode()

	// Log without secrets (no key here anyway).
	log.Info("ipapi resolve ip", "ip", ip, "url", u.Path+"?fields="+fields)

	headers := map[string]string{"Accept": "application/json"}

	// url prefix must start with "/"
	prefix := u.RequestURI()
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}

	// GET request; no body.
	sender := r.getNewSendRequestObject(nil, headers, prefix)
	if sender == nil {
		return resp, fmt.Errorf("ipapi_resolve_ip: internal sender error")
	}

	var raw RawResponse
	status, body, err := sender.SendRequest(&raw)
	if err != nil {
		log.Error("ipapi resolve ip", "status", status, "err", err.Error())
		return resp, err
	}
	_ = body // keep logs quiet

	// ip-api returns HTTP 200 even on provider failure; check raw.Status
	if raw.Status != "success" {
		if raw.Message == "" {
			raw.Message = "ip-api provider returned fail"
		}
		return resp, fmt.Errorf("ipapi_resolve_ip: provider error: %s", raw.Message)
	}

	// Map ip-api → unified struct
	resp.Ip = raw.Query
	resp.Type = "" // ip-api doesn't return type; leave empty or infer by simple heuristic if you need.
	resp.ContinentCode = raw.ContinentCode
	resp.ContinentName = raw.Continent
	resp.CountryCode = raw.CountryCode
	resp.CountryName = raw.Country
	resp.RegionCode = raw.Region
	resp.RegionName = raw.RegionName
	resp.City = raw.City
	resp.Zip = raw.Zip
	resp.Latitude = raw.Lat
	resp.Longitude = raw.Lon

	return resp, nil
}
