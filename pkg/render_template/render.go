package rendertemplate

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
)

type RenderTextStruct struct {
	TemplateDir  string
	TemplateName string
	Data         map[string]any
}

type RenderHTMLStruct struct {
	TemplateDir  string
	TemplateName string
	Data         map[string]any
}

func RenderTextTemplate(in RenderTextStruct) (string, error) {
	if strings.TrimSpace(in.TemplateName) == "" {
		return "", fmt.Errorf("empty template name")
	}
	path := filepath.Join(in.TemplateDir, fmt.Sprintf("%s.txt", in.TemplateName))

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
	if err := tpl.Execute(&sb, in.Data); err != nil {
		return "", fmt.Errorf("execute txt template: %w", err)
	}
	return sb.String(), nil
}

func RenderHTMLTemplate(in RenderHTMLStruct) (string, error) {
	if strings.TrimSpace(in.TemplateName) == "" {
		return "", fmt.Errorf("empty template name")
	}
	path := filepath.Join(in.TemplateDir, fmt.Sprintf("%s.html", in.TemplateName))

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
	if err := tpl.Execute(&sb, in.Data); err != nil {
		return "", fmt.Errorf("execute html template: %w", err)
	}
	return sb.String(), nil
}
