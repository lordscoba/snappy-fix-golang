package paystack

type InitRequest struct {
	Amount      int64                  `json:"amount"` // in kobo
	Email       string                 `json:"email"`
	Reference   string                 `json:"reference"`
	CallbackURL string                 `json:"callback_url,omitempty"`
	Currency    string                 `json:"currency,omitempty"` // "NGN"
	Channels    []string               `json:"channels,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`

	// NEW (either/or):
	Subaccount string     `json:"subaccount,omitempty"` // e.g. "ACCT_xxx"
	SplitCode  string     `json:"split_code,omitempty"` // e.g. "SPL_xxx"
	Split      *InitSplit `json:"split,omitempty"`      // optional, one-off inline split

	// Optional fee bearing (when using subaccount style)
	Bearer            string `json:"bearer,omitempty"`             // "account" or "subaccount"
	BearerSubaccount  string `json:"bearer_subaccount,omitempty"`  // ACCT_... when bearer=subaccount
	TransactionCharge int64  `json:"transaction_charge,omitempty"` // in kobo (flat)
}

// Inline split structure if you don’t want to pre-create a split object
type InitSplit struct {
	Type             string       `json:"type"`                        // "percentage" or "flat"
	Currency         string       `json:"currency,omitempty"`          // "NGN"
	Subaccounts      []SplitEntry `json:"subaccounts"`                 // [{subaccount: "ACCT_xxx", share: 20}, ...]
	BearerType       string       `json:"bearer_type,omitempty"`       // "account" | "subaccount"
	BearerSubaccount string       `json:"bearer_subaccount,omitempty"` // ACCT_... when bearer is subaccount
}

type SplitEntry struct {
	Subaccount string  `json:"subaccount"`
	Share      float64 `json:"share"` // percentage for "percentage" type; amount for "flat" type
}

type InitResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

// Verify Transaction
type VerifyResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		ID        int64  `json:"id"`
		Status    string `json:"status"` // "success"
		Reference string `json:"reference"`
		Amount    int64  `json:"amount"` // kobo
		Currency  string `json:"currency"`
		Channel   string `json:"channel"` // card/bank/ussd
		Customer  struct {
			ID           int64  `json:"id"`
			Email        string `json:"email"`
			CustomerCode string `json:"customer_code"`
		} `json:"customer"`
		Authorization struct {
			AuthorizationCode string `json:"authorization_code"`
			CardType          string `json:"card_type"`
			Last4             string `json:"last4"`
			ExpMonth          string `json:"exp_month"`
			ExpYear           string `json:"exp_year"`
			Reusable          bool   `json:"reusable"`
			Signature         string `json:"signature"`
		} `json:"authorization"`
	} `json:"data"`
}

// Create Transfer Recipient
type CreateRecipientRequest struct {
	Type          string `json:"type"` // "nuban"
	Name          string `json:"name"` // account name
	AccountNumber string `json:"account_number"`
	BankCode      string `json:"bank_code"`          // e.g. "058" (GTBank)
	Currency      string `json:"currency,omitempty"` // "NGN"
}

type CreateRecipientResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		RecipientCode string `json:"recipient_code"`
	} `json:"data"`
}

// Transfer
type TransferRequest struct {
	Source    string `json:"source"`    // "balance"
	Amount    int64  `json:"amount"`    // kobo
	Recipient string `json:"recipient"` // recipient_code
	Reason    string `json:"reason,omitempty"`
	Reference string `json:"reference,omitempty"` // your unique payout ref
	Currency  string `json:"currency,omitempty"`  // NGN
}

type TransferResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		TransferCode string `json:"transfer_code"`
		Status       string `json:"status"`
		ID           int64  `json:"id"`
		Reference    string `json:"reference"`
	} `json:"data"`
}

// --- List Banks (types) ---
type ListBanksRequest struct {
	Country  string `json:"country,omitempty"`  // e.g. "nigeria"
	Currency string `json:"currency,omitempty"` // e.g. "NGN"
	Type     string `json:"type,omitempty"`     // e.g. "nuban"
	PerPage  int    `json:"per_page,omitempty"` // optional
	Page     int    `json:"page,omitempty"`     // optional
}

type Bank struct {
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Code     string `json:"code"`     // <-- bank_code used for transfers
	Longcode string `json:"longcode"` // optional
	Active   bool   `json:"active"`
	Country  string `json:"country"`
	Currency string `json:"currency"`
	Type     string `json:"type"`
}

type ListBanksResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    []Bank `json:"data"`
	// Paystack may include Meta; we omit it since it's not required by our API
}

type CreateSubaccountRequest struct {
	BusinessName     string  `json:"business_name" binding:"required"`
	SettlementBank   string  `json:"settlement_bank" binding:"required"` // bank code, e.g. "011"
	AccountNumber    string  `json:"account_number" binding:"required"`
	PercentageCharge float64 `json:"percentage_charge,omitempty"` // e.g. 20.0 for 20%
	Description      string  `json:"description,omitempty"`
	Currency         string  `json:"currency,omitempty"` // "NGN" (defaults to merchant’s currency)
}

type CreateSubaccountResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		SubaccountCode string  `json:"subaccount_code"` // ACCT_xxx
		BusinessName   string  `json:"business_name"`
		SettlementBank string  `json:"settlement_bank"`
		AccountNumber  string  `json:"account_number"`
		Percentage     float64 `json:"percentage_charge"`
	} `json:"data"`
}

// ===== Create Split (pre-defined multi-beneficiary) =====
type CreateSplitRequest struct {
	Name             string       `json:"name" binding:"required"`
	Type             string       `json:"type" binding:"required"`     // "percentage" | "flat"
	Currency         string       `json:"currency" binding:"required"` // "NGN"
	Subaccounts      []SplitEntry `json:"subaccounts" binding:"required"`
	BearerType       string       `json:"bearer_type,omitempty"`       // "account" | "subaccount"
	BearerSubaccount string       `json:"bearer_subaccount,omitempty"` // ACCT_xxx
}

type CreateSplitResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		SplitCode string `json:"split_code"` // SPL_xxx
		// other fields omitted
	} `json:"data"`
}
