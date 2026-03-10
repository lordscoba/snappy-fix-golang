package mailgun

type SendEmailRequest struct {
	From    string   // "Sender <sender@yourdomain.com>"
	To      []string // one or more recipients
	Subject string
	Text    string // plain text body
	HTML    string // optional HTML
	// You can add CC, BCC, attachments, etc., as needed.
}

type SendEmailResponse struct {
	ID      string `json:"id,omitempty"`
	Message string `json:"message,omitempty"`
}
