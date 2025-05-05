package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"path/filepath"

	"gopkg.in/mail.v2"
)

type Mailer struct {
	dialer    *mail.Dialer
	templates *template.Template
	from      string
}

type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	TemplateDir string
}

func New(config Config) (*Mailer, error) {
	dialer := mail.NewDialer(config.Host, config.Port, config.Username, config.Password)
	dialer.SSL = true

	s, err := dialer.Dial()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	s.Close()

	templates, err := template.ParseGlob(filepath.Join(config.TemplateDir, "*.html"))
	if err != nil {
		return nil, fmt.Errorf("failed to load email templates: %w", err)
	}

	return &Mailer{
		dialer:    dialer,
		templates: templates,
		from:      config.From,
	}, nil
}

func (m *Mailer) Send(to, subject, templateName string, data interface{}) error {
	tmpl := m.templates.Lookup(templateName)
	if tmpl == nil {
		return fmt.Errorf("template %s not found", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, templateName, data); err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	msg := mail.NewMessage()
	msg.SetHeader("From", m.from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", buf.String())
	if err := m.dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

func (m *Mailer) SendWithAttachments(to, subject, templateName string, data interface{}, attachments []string) error {
	tmpl := m.templates.Lookup(templateName)
	if tmpl == nil {
		return fmt.Errorf("template %s not found", templateName)
	}

	var buf bytes.Buffer
	if err := tmpl.ExecuteTemplate(&buf, templateName, data); err != nil {
		return fmt.Errorf("failed to render template: %w", err)
	}

	msg := mail.NewMessage()
	msg.SetHeader("From", m.from)
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", buf.String())

	for _, attachment := range attachments {
		msg.Attach(attachment)
	}

	if err := m.dialer.DialAndSend(msg); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
} 
