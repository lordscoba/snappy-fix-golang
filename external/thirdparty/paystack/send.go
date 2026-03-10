package paystack

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"net/url"
	"path"

	"github.com/snappy-fix-golang/internal/config"
)

// ---------- Methods ----------

func (r *RequestObj) InitializePayment() (InitResponse, error) {
	var (
		cfg  = config.GetConfig()
		resp InitResponse
		log  = r.Logger
	)
	in, ok := r.RequestData.(InitRequest)
	if !ok {
		return resp, fmt.Errorf("paystack_init: bad request type")
	}

	base, err := url.Parse(r.Path)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return resp, fmt.Errorf("paystack_init: invalid base url")
	}
	prefix := path.Join("/", "transaction", "initialize")

	headers := map[string]string{
		"Authorization": "Bearer " + cfg.Paystack.PaystackSecretKey,
		"Accept":        "application/json",
	}

	s := r.sender(in, headers, prefix)
	if s == nil {
		return resp, fmt.Errorf("paystack_init: sender nil")
	}
	status, _, err := s.SendRequest(&resp)
	if err != nil {
		log.Error("paystack init", "status", status, "err", err.Error())
		return resp, err
	}
	if !resp.Status {
		return resp, fmt.Errorf("paystack_init: %s", resp.Message)
	}
	return resp, nil
}

func (r *RequestObj) VerifyPayment(reference string) (VerifyResponse, error) {
	var (
		cfg  = config.GetConfig()
		resp VerifyResponse
		log  = r.Logger
	)
	if reference == "" {
		return resp, fmt.Errorf("paystack_verify: empty reference")
	}
	base, err := url.Parse(r.Path)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return resp, fmt.Errorf("paystack_verify: invalid base url")
	}
	prefix := path.Join("/", "transaction", "verify", reference)

	headers := map[string]string{
		"Authorization": "Bearer " + cfg.Paystack.PaystackSecretKey,
		"Accept":        "application/json",
	}

	// GET with nil body
	s := r.sender(nil, headers, prefix)
	if s == nil {
		return resp, fmt.Errorf("paystack_verify: sender nil")
	}
	// Override method/status for GET
	s.Method = "GET"
	s.SuccessCode = 200

	status, _, err := s.SendRequest(&resp)
	if err != nil {
		log.Error("paystack verify", "status", status, "err", err.Error())
		return resp, err
	}
	if !resp.Status {
		return resp, fmt.Errorf("paystack_verify: %s", resp.Message)
	}
	return resp, nil
}

func (r *RequestObj) CreateTransferRecipient(req CreateRecipientRequest) (CreateRecipientResponse, error) {
	var (
		cfg  = config.GetConfig()
		resp CreateRecipientResponse
		log  = r.Logger
	)
	base, err := url.Parse(r.Path)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return resp, fmt.Errorf("paystack_create_recipient: invalid base url")
	}
	prefix := path.Join("/", "transferrecipient")

	headers := map[string]string{
		"Authorization": "Bearer " + cfg.Paystack.PaystackSecretKey,
		"Accept":        "application/json",
	}

	s := r.sender(req, headers, prefix)
	if s == nil {
		return resp, fmt.Errorf("paystack_create_recipient: sender nil")
	}
	fmt.Println(s)
	status, _, err := s.SendRequest(&resp)
	if err != nil {
		log.Error("paystack create recipient", "status", status, "err", err.Error())
		return resp, err
	}
	if !resp.Status {
		return resp, fmt.Errorf("paystack_create_recipient: %s", resp.Message)
	}
	return resp, nil
}

func (r *RequestObj) Transfer(req TransferRequest) (TransferResponse, error) {
	var (
		cfg  = config.GetConfig()
		resp TransferResponse
		log  = r.Logger
	)
	base, err := url.Parse(r.Path)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return resp, fmt.Errorf("paystack_transfer: invalid base url")
	}
	prefix := path.Join("/", "transfer")

	headers := map[string]string{
		"Authorization": "Bearer " + cfg.Paystack.PaystackSecretKey,
		"Accept":        "application/json",
	}

	s := r.sender(req, headers, prefix)
	if s == nil {
		return resp, fmt.Errorf("paystack_transfer: sender nil")
	}
	status, _, err := s.SendRequest(&resp)
	if err != nil {
		log.Error("paystack transfer", "status", status, "err", err.Error())
		return resp, err
	}
	if !resp.Status {
		return resp, fmt.Errorf("paystack_transfer: %s", resp.Message)
	}
	return resp, nil
}

func (r *RequestObj) ListBanks(req ListBanksRequest) (ListBanksResponse, error) {
	var (
		cfg  = config.GetConfig()
		resp ListBanksResponse
		log  = r.Logger
	)

	base, err := url.Parse(r.Path)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return resp, fmt.Errorf("paystack_list_banks: invalid base url")
	}

	// Build /bank?country=...&currency=...&type=...&perPage=&page=
	q := url.Values{}
	if req.Country != "" {
		q.Set("country", req.Country)
	}
	if req.Currency != "" {
		q.Set("currency", req.Currency)
	}
	if req.Type != "" {
		q.Set("type", req.Type)
	}
	if req.PerPage > 0 {
		q.Set("perPage", fmt.Sprintf("%d", req.PerPage))
	}
	if req.Page > 0 {
		q.Set("page", fmt.Sprintf("%d", req.Page))
	}

	// /bank + query
	prefix := path.Join("/", "bank")
	if qs := q.Encode(); qs != "" {
		prefix = prefix + "?" + qs
	}

	headers := map[string]string{
		"Authorization": "Bearer " + cfg.Paystack.PaystackSecretKey,
		"Accept":        "application/json",
	}

	// GET with nil body
	s := r.sender(nil, headers, prefix)
	if s == nil {
		return resp, fmt.Errorf("paystack_list_banks: sender nil")
	}
	s.Method = "GET"
	s.SuccessCode = 200

	status, _, err := s.SendRequest(&resp)
	if err != nil {
		log.Error("paystack list banks", "status", status, "err", err.Error())
		return resp, err
	}
	if !resp.Status {
		return resp, fmt.Errorf("paystack_list_banks: %s", resp.Message)
	}
	return resp, nil
}

// CreateSubaccount: POST /subaccount  (201 Created)
func (r *RequestObj) CreateSubaccount(req CreateSubaccountRequest) (CreateSubaccountResponse, error) {
	var (
		cfg  = config.GetConfig()
		resp CreateSubaccountResponse
		log  = r.Logger
	)
	base, err := url.Parse(r.Path)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return resp, fmt.Errorf("paystack_create_subaccount: invalid base url")
	}
	prefix := path.Join("/", "subaccount")
	headers := map[string]string{
		"Authorization": "Bearer " + cfg.Paystack.PaystackSecretKey,
		"Accept":        "application/json",
	}

	s := r.sender(req, headers, prefix)
	if s == nil {
		return resp, fmt.Errorf("paystack_create_subaccount: sender nil")
	}
	s.SuccessCode = 201
	status, _, err := s.SendRequest(&resp)
	if err != nil {
		log.Error("paystack create subaccount", "status", status, "err", err.Error())
		return resp, err
	}
	if !resp.Status {
		return resp, fmt.Errorf("paystack_create_subaccount: %s", resp.Message)
	}
	return resp, nil
}

// CreateSplit: POST /split  (201 Created)
func (r *RequestObj) CreateSplit(req CreateSplitRequest) (CreateSplitResponse, error) {
	var (
		cfg  = config.GetConfig()
		resp CreateSplitResponse
		log  = r.Logger
	)
	base, err := url.Parse(r.Path)
	if err != nil || base.Scheme == "" || base.Host == "" {
		return resp, fmt.Errorf("paystack_create_split: invalid base url")
	}
	prefix := path.Join("/", "split")
	headers := map[string]string{
		"Authorization": "Bearer " + cfg.Paystack.PaystackSecretKey,
		"Accept":        "application/json",
	}

	s := r.sender(req, headers, prefix)
	if s == nil {
		return resp, fmt.Errorf("paystack_create_split: sender nil")
	}
	s.SuccessCode = 201
	status, _, err := s.SendRequest(&resp)
	if err != nil {
		log.Error("paystack create split", "status", status, "err", err.Error())
		return resp, err
	}
	if !resp.Status {
		return resp, fmt.Errorf("paystack_create_split: %s", resp.Message)
	}
	return resp, nil
}

// ---------- Webhook Signature Helper (optional) ----------
// Paystack sends `X-Paystack-Signature` = HMAC-SHA512(body, SECRET_KEY)
func VerifyWebhookSignature(secretKey string, body []byte, signature string) bool {
	mac := hmac.New(sha512.New, []byte(secretKey))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	// Compare case-insensitively
	return hmac.Equal([]byte(expected), []byte(signature))
}
