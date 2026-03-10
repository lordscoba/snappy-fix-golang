package request

import (
	"fmt"

	"github.com/snappy-fix-golang/external"
	"github.com/snappy-fix-golang/external/mocks"
	"github.com/snappy-fix-golang/external/thirdparty/paystack"
	"github.com/snappy-fix-golang/internal/config"
)

func (er ExternalRequest) SendExternalRequestPaystack(name string, data interface{}) (interface{}, error) {
	cfg := config.GetConfig()

	if !er.Test {
		switch name {
		// ---------- Paystack ----------
		case external.PaystackInit:
			obj := paystack.RequestObj{
				Name:         name,
				Path:         fmt.Sprintf("%v", cfg.Paystack.PaystackBaseUrl),
				Method:       "POST",
				SuccessCode:  200,
				DecodeMethod: JsonDecodeMethod,
				RequestData:  data, // paystack.InitRequest
				Logger:       er.Logger,
			}
			return obj.InitializePayment()

		case external.PaystackVerify:
			ref, ok := data.(string)
			if !ok {
				return nil, fmt.Errorf("paystack_verify: data must be reference string")
			}
			obj := paystack.RequestObj{
				Name:         name,
				Path:         fmt.Sprintf("%v", cfg.Paystack.PaystackBaseUrl),
				Method:       "GET", // overridden in method
				SuccessCode:  200,
				DecodeMethod: JsonDecodeMethod,
				RequestData:  nil,
				Logger:       er.Logger,
			}
			return obj.VerifyPayment(ref)

		case external.PaystackCreateRecipient:
			obj := paystack.RequestObj{
				Name:         name,
				Path:         fmt.Sprintf("%v", cfg.Paystack.PaystackBaseUrl),
				Method:       "POST",
				SuccessCode:  201,
				DecodeMethod: JsonDecodeMethod,
				RequestData:  data, // paystack.CreateRecipientRequest
				Logger:       er.Logger,
			}
			req, ok := data.(paystack.CreateRecipientRequest)
			if !ok {
				return nil, fmt.Errorf("paystack_create_recipient: bad request type")
			}
			return obj.CreateTransferRecipient(req)

		case external.PaystackTransfer:
			obj := paystack.RequestObj{
				Name:         name,
				Path:         fmt.Sprintf("%v", cfg.Paystack.PaystackBaseUrl),
				Method:       "POST",
				SuccessCode:  200,
				DecodeMethod: JsonDecodeMethod,
				RequestData:  data, // paystack.TransferRequest
				Logger:       er.Logger,
			}
			req, ok := data.(paystack.TransferRequest)
			if !ok {
				return nil, fmt.Errorf("paystack_transfer: bad request type")
			}
			return obj.Transfer(req)

		case external.PaystackListBanks:
			obj := paystack.RequestObj{
				Name:         name,
				Path:         fmt.Sprintf("%v", cfg.Paystack.PaystackBaseUrl),
				Method:       "GET",
				SuccessCode:  200,
				DecodeMethod: JsonDecodeMethod,
				RequestData:  nil,
				Logger:       er.Logger,
			}
			req, ok := data.(paystack.ListBanksRequest)
			if !ok {
				return nil, fmt.Errorf("paystack_list_banks: bad request type")
			}
			return obj.ListBanks(req)

		case external.PaystackCreateSubaccount:
			obj := paystack.RequestObj{
				Name:         name,
				Path:         fmt.Sprintf("%v", cfg.Paystack.PaystackBaseUrl),
				Method:       "POST",
				SuccessCode:  201,
				DecodeMethod: JsonDecodeMethod,
				RequestData:  data, // paystack.CreateSubaccountRequest
				Logger:       er.Logger,
			}
			req, ok := data.(paystack.CreateSubaccountRequest)
			if !ok {
				return nil, fmt.Errorf("paystack_create_subaccount: bad request type")
			}
			return obj.CreateSubaccount(req)

		case external.PaystackCreateSplit:
			obj := paystack.RequestObj{
				Name:         name,
				Path:         fmt.Sprintf("%v", cfg.Paystack.PaystackBaseUrl),
				Method:       "POST",
				SuccessCode:  201,
				DecodeMethod: JsonDecodeMethod,
				RequestData:  data, // paystack.CreateSplitRequest
				Logger:       er.Logger,
			}
			req, ok := data.(paystack.CreateSplitRequest)
			if !ok {
				return nil, fmt.Errorf("paystack_create_split: bad request type")
			}
			return obj.CreateSplit(req)
		default:
			return nil, fmt.Errorf("request: unsupported request name: %s", name)
		}
	}

	// Test path: delegate to mocks
	mer := mocks.ExternalRequest{Logger: er.Logger, Test: true}
	return mer.SendExternalRequest(name, data)
}
