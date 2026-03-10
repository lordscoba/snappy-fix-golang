package pingservices

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/snappy-fix-golang/external"
	"github.com/snappy-fix-golang/external/request"
	"github.com/snappy-fix-golang/external/thirdparty/mailgun"
	logutil "github.com/snappy-fix-golang/pkg/logger"
)

// Service owns notification business logic (email + push).
type Service struct {
	ExtReq request.ExternalRequest
	Logger *logutil.Logger

	// Base dir for templates, e.g. "./templates/email"
	TemplateDir string
}

// ---------- Email (Mailgun) ----------

// EmailSendInput defines a flexible, template-powered email.
type EmailSendInput struct {
	From         string   // e.g. "Your App <no-reply@mg.yourdomain.com>"
	To           []string // ["user@example.com", ...]
	Subject      string
	Plaintext    bool           // if true, use *.txt template; else *.html
	Template     string         // logical name, e.g. "welcome" => welcome.html / welcome.txt
	Vars         map[string]any // variables for templating (Name, Email, LogoURL, WebsiteURL, etc.)
	FallbackTXT  string         // optional fallback plain text content if template missing
	FallbackHTML string         // optional fallback HTML content if template missing
}

// SendEmail renders a template (HTML or TXT) and sends via Mailgun.
func (s Service) SendEmail(in EmailSendInput) (mailgun.SendEmailResponse, error) {
	var out mailgun.SendEmailResponse

	// Basic validation here; handlers already validate, but double-check critical parts.
	if len(in.To) == 0 {
		return out, fmt.Errorf("email: at least one recipient required")
	}
	if strings.TrimSpace(in.From) == "" {
		return out, fmt.Errorf("email: from is required")
	}
	if strings.TrimSpace(in.Subject) == "" {
		return out, fmt.Errorf("email: subject is required")
	}
	if strings.TrimSpace(in.Template) == "" && !in.Plaintext && in.FallbackHTML == "" {
		return out, fmt.Errorf("email: template name or FallbackHTML required for HTML")
	}
	if strings.TrimSpace(in.Template) == "" && in.Plaintext && in.FallbackTXT == "" {
		return out, fmt.Errorf("email: template name or FallbackTXT required for TXT")
	}

	// Render body
	var textBody, htmlBody string
	var err error

	if in.Plaintext {
		textBody, err = s.renderTextTemplate(in.Template, in.Vars)
		if err != nil {
			if in.FallbackTXT == "" {
				return out, fmt.Errorf("email: render txt template: %w", err)
			}
			textBody = in.FallbackTXT
		}
	} else {
		htmlBody, err = s.renderHTMLTemplate(in.Template, in.Vars)
		if err != nil {
			if in.FallbackHTML == "" {
				return out, fmt.Errorf("email: render html template: %w", err)
			}
			htmlBody = in.FallbackHTML
		}
	}

	req := mailgun.SendEmailRequest{
		From:    in.From,
		To:      in.To,
		Subject: in.Subject,
		Text:    textBody,
		HTML:    htmlBody,
	}

	respAny, err := s.ExtReq.SendExternalRequestMailgun(external.MailgunSend, req)
	if err != nil {
		fmt.Println(err)
		return out, fmt.Errorf("email: send mailgun failed: %w", err)
	}

	typed, ok := respAny.(mailgun.SendEmailResponse)
	if !ok {
		return out, fmt.Errorf("email: unexpected response type from mailgun")
	}

	return typed, nil
}

// renderHTMLTemplate loads ./templates/email/<name>.html and executes it with data.
func (s Service) renderHTMLTemplate(name string, data map[string]any) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("empty template name")
	}
	path := filepath.Join(s.TemplateDir, fmt.Sprintf("%s.html", name))

	tplBytes, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read html template: %w", err)
	}
	// html/template auto-escapes. Your variables become {{.Name}}, {{.Email}}, {{.LogoURL}} etc.
	tpl, err := template.New(filepath.Base(path)).Parse(string(tplBytes))
	if err != nil {
		return "", fmt.Errorf("parse html template: %w", err)
	}
	var sb strings.Builder
	if err := tpl.Execute(&sb, data); err != nil {
		return "", fmt.Errorf("execute html template: %w", err)
	}
	return sb.String(), nil
}

// renderTextTemplate loads ./templates/email/<name>.txt and executes it with data.
func (s Service) renderTextTemplate(name string, data map[string]any) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("empty template name")
	}
	path := filepath.Join(s.TemplateDir, fmt.Sprintf("%s.txt", name))

	tplBytes, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read txt template: %w", err)
	}
	// Plain text: use text/template-like parsing via html/template (safe for text too).
	tpl, err := template.New(filepath.Base(path)).Parse(string(tplBytes))
	if err != nil {
		return "", fmt.Errorf("parse txt template: %w", err)
	}
	var sb strings.Builder
	if err := tpl.Execute(&sb, data); err != nil {
		return "", fmt.Errorf("execute txt template: %w", err)
	}
	return sb.String(), nil
}

// ---------- Push (FCM) ----------

type PushSendInput struct {
	Token string // OR Topic
	Topic string
	Title string
	Body  string
	Data  map[string]string
}
