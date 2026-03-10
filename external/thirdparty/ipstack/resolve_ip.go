package ipstack

import (
	"fmt"
	"net/url"
	"path"
	"strings"

	"github.com/snappy-fix-golang/internal/config"
)

func (r *RequestObj) IpstackResolveIp() (IPStackResolveIPResponse, error) {
	var (
		key  = config.GetConfig().IPStack.Key
		resp IPStackResolveIPResponse
		log  = r.Logger
		in   = r.RequestData
	)

	ip, ok := in.(string)
	if !ok {
		log.Error("ipstack resolve ip", in, "request data format error: expected string IP")
		return resp, fmt.Errorf("ipstack_resolve_ip: request data format error (expected string IP)")
	}

	// Build URL safely: base + /{ip} + ?access_key=...
	u, err := url.Parse(r.Path)
	if err != nil {
		log.Error("ipstack resolve ip", "invalid base url", r.Path, err.Error())
		return resp, fmt.Errorf("ipstack_resolve_ip: invalid base URL")
	}
	u.Path = path.Join(u.Path, ip)

	q := u.Query()
	q.Set("access_key", key) // IPStack requires the key as a query parameter.
	u.RawQuery = q.Encode()

	// Never log the full URL (it contains the key). Log a masked version.
	masked := u.String()
	if qi := u.Query().Get("access_key"); qi != "" {
		masked = strings.Replace(masked, qi, "****", 1)
	}
	log.Info("ipstack resolve ip", "ip", ip, "url", masked)

	// Prepare headers (explicitly accept JSON).
	headers := map[string]string{
		"Accept": "application/json",
	}

	// Build the URL prefix that will be appended to base Path.
	// Use RequestURI(), but ensure it starts with a leading slash.
	prefix := u.RequestURI()
	if !strings.HasPrefix(prefix, "/") {
		prefix = "/" + prefix // ensure leading slash
	}

	// GET request; no body.
	sender := r.getNewSendRequestObject(nil, headers, prefix)
	if sender == nil {
		return resp, fmt.Errorf("ipstack_resolve_ip: internal sender error")
	}

	// Execute
	status, body, err := sender.SendRequest(&resp)
	if err != nil {
		log.Error("ipstack resolve ip", "status", status, "err", err.Error())
		return resp, err
	}

	_ = body // keep logs clean

	return resp, nil
}
