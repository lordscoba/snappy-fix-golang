package mailgun

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"path"

	"github.com/snappy-fix-golang/internal/config"
)

func (r *RequestObj) MailgunSend() (SendEmailResponse, error) {
	var (
		cfg  = config.GetConfig()
		resp SendEmailResponse
		log  = r.Logger
	)

	in, ok := r.RequestData.(SendEmailRequest)
	if !ok {
		log.Error("mailgun send", r.RequestData, "request data format error: expected mailgun.SendEmailRequest")
		return resp, fmt.Errorf("mailgun_send: request data format error (expected mailgun.SendEmailRequest)")
	}

	// r.Path should be https://api.mailgun.net/v3  (with scheme!)
	base, err := url.Parse(r.Path)
	if err != nil || base.Scheme == "" || base.Host == "" {
		log.Error("mailgun send", "invalid base url", r.Path)
		return resp, fmt.Errorf("mailgun_send: invalid base URL")
	}

	// Prefix must be only "/<domain>/messages"
	prefix := path.Join("/", cfg.Mailgun.Domain, "messages")

	// Build form body
	form := url.Values{}
	form.Set("from", in.From)
	for _, to := range in.To {
		form.Add("to", to)
	}
	form.Set("subject", in.Subject)
	if in.Text != "" {
		form.Set("text", in.Text)
	}
	if in.HTML != "" {
		form.Set("html", in.HTML)
	}
	bodyBytes := []byte(form.Encode())

	// Basic auth: api:<key>
	basic := "Basic " + base64.StdEncoding.EncodeToString([]byte("api:"+cfg.Mailgun.ApiKey))

	headers := map[string]string{
		"Accept":        "application/json",
		"Authorization": basic,
		// Content-Type set below via override
	}

	sender := r.getNewSendRequestObject(nil, headers, prefix)
	if sender == nil {
		return resp, fmt.Errorf("mailgun_send: internal sender error")
	}

	// These fields MUST exist in your transport struct & logic:
	sender.BodyBytes = bodyBytes                                     // <--- add this field to transport
	sender.ContentTypeOverride = "application/x-www-form-urlencoded" // <--- add this too

	status, raw, err := sender.SendRequest(&resp)
	if err != nil {
		log.Error("mailgun send", "status", status, "err", err.Error())
		return resp, err
	}
	_ = raw
	return resp, nil
}
