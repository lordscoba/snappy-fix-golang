package request

import (
	"fmt"

	"github.com/snappy-fix-golang/external"
	"github.com/snappy-fix-golang/external/mocks"
	"github.com/snappy-fix-golang/external/thirdparty/mailgun"
	"github.com/snappy-fix-golang/internal/config"
)

func (er ExternalRequest) SendExternalRequestMailgun(name string, data interface{}) (interface{}, error) {
	cfg := config.GetConfig()

	if !er.Test {
		switch name {
		// ---------- mailgun operations ----------
		case external.MailgunSend:
			obj := mailgun.RequestObj{
				Name:         name,
				Path:         fmt.Sprintf("%v", cfg.Mailgun.BaseUrl), // https://api.mailgun.net/v3
				Method:       "POST",
				SuccessCode:  200,
				DecodeMethod: JsonDecodeMethod, // response is JSON
				RequestData:  data,             // mailgun.SendEmailRequest
				Logger:       er.Logger,
			}
			return obj.MailgunSend()
		default:
			return nil, fmt.Errorf("request: unsupported request name: %s", name)
		}
	}

	// Test path: delegate to mocks
	mer := mocks.ExternalRequest{Logger: er.Logger, Test: true}
	return mer.SendExternalRequest(name, data)
}
