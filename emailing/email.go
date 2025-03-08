package emailing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"path/filepath"
	"runtime"
	"time"

	"github.com/spf13/viper"
	"go.uber.org/zap"
	gomail "gopkg.in/mail.v2"
)

type EmailService struct {
	smtpHost string
	smtpPort int
	from     string
	password string
	tmpl     *template.Template
	dialer   *gomail.Dialer
	logger   *zap.Logger
}

type TextSection struct {
	Text string `json:"text"`
}

type Button struct {
	Text    string `json:"text"`
	URL     string `json:"url"`
	APIKey  string `json:"apiKey,omitempty"`
	Payload string `json:"payload,omitempty"`
}

type Attachment struct {
	Filename    string
	Content     []byte
	ContentType string
}

type EmailOptions struct {
	To          string
	Subject     string
	Sections    []TextSection
	Buttons     []Button
	ExpiryTime  time.Duration
	Attachments []Attachment
}

func CreateSimpleButton(text, url string) Button {
	return Button{
		Text: text,
		URL:  url,
	}
}

func CreateAdvancedButton(text, link, apiKey string, payload interface{}) (Button, error) {
	baseURL, err := url.Parse(link)
	if err != nil {
		return Button{}, fmt.Errorf("invalid URL: %w", err)
	}

	query := baseURL.Query()
	query.Set("api_key", apiKey)

	if payload != nil {
		payloadMap := make(map[string]interface{})
		if pm, ok := payload.(map[string]interface{}); ok {
			payloadMap = pm
		} else {
			payloadBytes, er := json.Marshal(payload)
			if er != nil {
				return Button{}, fmt.Errorf("failed to marshal payload: %w", err)
			}
			if err = json.Unmarshal(payloadBytes, &payloadMap); err != nil {
				return Button{}, fmt.Errorf("failed to convert payload to map: %w", err)
			}
		}

		for key, value := range payloadMap {
			strValue := fmt.Sprintf("%v", value)
			query.Set(key, strValue)
		}
	}

	baseURL.RawQuery = query.Encode()

	button := Button{
		Text: text,
		URL:  baseURL.String(),
	}

	if button.URL == "" {
		return Button{}, fmt.Errorf("URL cannot be empty for advanced button")
	}

	return button, nil
}

type EmailTemplateData struct {
	LogoURL    string
	Subject    string
	Sections   []TextSection
	Buttons    []Button
	ExpiryTime string
}

func New(cfg *viper.Viper, logger *zap.Logger) (*EmailService, error) {
	if cfg == nil {
		return nil, fmt.Errorf("emailing config instance is nil")
	}

	service := &EmailService{
		smtpHost: cfg.GetString("email.smtpHost"),
		smtpPort: cfg.GetInt("email.smtpPort"),
		from:     cfg.GetString("email.from"),
		password: cfg.GetString("email.pwd"),
		logger:   logger,
	}

	if service.smtpHost == "" || service.smtpPort == 0 || service.from == "" ||
		service.password == "" {
		return nil, fmt.Errorf("missing required email configuration")
	}

	_, currentFile, _, _ := runtime.Caller(0)
	dir := filepath.Dir(currentFile)

	tmplPath := filepath.Join(dir, "email_template.html")

	tmpl, err := template.ParseFiles(tmplPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse email template: %w", err)
	}

	service.tmpl = tmpl
	service.dialer = gomail.NewDialer(
		service.smtpHost,
		service.smtpPort,
		service.from,
		service.password,
	)
	service.dialer.StartTLSPolicy = gomail.MandatoryStartTLS

	return service, nil
}

func (s *EmailService) SendEmail(opts EmailOptions) error {
	if opts.To == "" {
		return fmt.Errorf("recipient email address is required")
	}
	if opts.Subject == "" {
		return fmt.Errorf("email subject is required")
	}

	var tplBuffer bytes.Buffer

	s.logger.Debug("Sending email",
		zap.String("to", opts.To),
		zap.String("subject", opts.Subject),
		zap.Int("num_sections", len(opts.Sections)),
		zap.Int("num_buttons", len(opts.Buttons)),
		zap.Int("num_attachments", len(opts.Attachments)))

	data := EmailTemplateData{
		LogoURL:  "https://github.com/sabouaram/wasselli_logo/blob/main/wasselli.png",
		Subject:  opts.Subject,
		Sections: opts.Sections,
		Buttons:  opts.Buttons,
	}

	if opts.ExpiryTime > 0 {
		data.ExpiryTime = formatDuration(opts.ExpiryTime)
	}

	if err := s.tmpl.Execute(&tplBuffer, data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	message := gomail.NewMessage()
	message.SetHeader("From", s.from)
	message.SetHeader("To", opts.To)
	message.SetHeader("Subject", opts.Subject)
	message.SetBody("text/html", tplBuffer.String())

	for _, attachment := range opts.Attachments {
		if attachment.Content == nil || len(attachment.Content) == 0 {
			s.logger.Warn("Skipping empty attachment",
				zap.String("filename", attachment.Filename))
			continue
		}

		reader := bytes.NewReader(attachment.Content)

		message.AttachReader(
			attachment.Filename,
			reader,
		)

		s.logger.Debug("Added attachment",
			zap.String("filename", attachment.Filename),
			zap.String("content_type", attachment.ContentType),
			zap.Int("size", len(attachment.Content)))
	}

	if err := s.dialer.DialAndSend(message); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	s.logger.Info("Email sent successfully",
		zap.String("to", opts.To),
		zap.String("subject", opts.Subject))

	return nil
}

func CreatePDFAttachment(filename string, content []byte) Attachment {
	return Attachment{
		Filename:    filename,
		Content:     content,
		ContentType: "application/pdf",
	}
}

func CreateAttachment(filename string, content []byte, contentType string) Attachment {
	return Attachment{
		Filename:    filename,
		Content:     content,
		ContentType: contentType,
	}
}

func formatDuration(d time.Duration) string {
	minutes := int(d.Minutes())

	if minutes < 60 {
		return fmt.Sprintf("%d minutes", minutes)
	}

	hours := minutes / 60
	remainingMinutes := minutes % 60

	if remainingMinutes == 0 {
		return fmt.Sprintf("%d hours", hours)
	}

	return fmt.Sprintf("%d hours and %d minutes", hours, remainingMinutes)
}
