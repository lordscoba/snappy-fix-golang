package ipdata

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/snappy-fix-golang/external/thirdparty/ipstack" // [REUSABLE-3]
	"github.com/snappy-fix-golang/internal/config"
)

// IpdataResolveIp calls ipdata.co and maps to the unified IPStackResolveIPResponse.
func (r *RequestObj) IpdataResolveIp() (ipstack.IPStackResolveIPResponse, error) {
	var (
		key  = config.GetConfig().IPData.Key // ensure config has IPData.{BaseUrl, Key}
		resp ipstack.IPStackResolveIPResponse
		log  = r.Logger
		in   = r.RequestData
	)

	ip, ok := in.(string)
	if !ok {
		log.Error("ipdata resolve ip", in, "request data format error: expected string IP")
		return resp, fmt.Errorf("ipdata_resolve_ip: request data format error (expected string IP)")
	}

	// Build URL: {BaseURL}/{ip}?api-key={key}
	u, err := url.Parse(r.Path)
	if err != nil {
		log.Error("ipdata resolve ip", "invalid base url", r.Path, err.Error())
		return resp, fmt.Errorf("ipdata_resolve_ip: invalid base URL")
	}
	u.Path = path.Join(u.Path, ip)
	q := u.Query()
	q.Set("api-key", key)
	u.RawQuery = q.Encode()

	// Mask key in logs
	masked := u.String()
	if ak := q.Get("api-key"); ak != "" {
		masked = strings.Replace(masked, ak, "****", 1)
	}
	log.Info("ipdata resolve ip", "ip", ip, "url", masked)

	headers := map[string]string{"Accept": "application/json"}

	// Ensure leading slash for UrlPrefix
	prefix := u.RequestURI()
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix
	}

	// GET with no body — [REUSABLE-1]
	sender := r.getNewSendRequestObject(nil, headers, prefix)
	if sender == nil {
		return resp, fmt.Errorf("ipdata_resolve_ip: internal sender error")
	}

	var raw RawResponse
	status, body, err := sender.SendRequest(&raw)
	if err != nil {
		log.Error("ipdata resolve ip", "status", status, "err", err.Error())
		return resp, err
	}
	_ = body // don’t log payloads

	// Map ipdata → unified struct
	resp.Ip = raw.IP
	resp.Type = raw.Type
	resp.ContinentCode = raw.ContinentCode
	resp.ContinentName = raw.ContinentName
	resp.CountryCode = raw.CountryCode
	resp.CountryName = raw.CountryName
	resp.RegionCode = raw.RegionCode
	resp.RegionName = raw.Region
	resp.City = raw.City
	resp.Zip = raw.Postal
	resp.Latitude = raw.Latitude
	resp.Longitude = raw.Longitude

	return resp, nil
}
