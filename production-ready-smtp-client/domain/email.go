package domain

import (
	"fmt"
	"regexp"
	"time"
)

// Email represents an email message
type Email struct {
	ID          string
	From        string
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	TextBody    string
	HTMLBody    string
	Attachments []Attachment
	Headers     map[string]string
	Priority    Priority
	CreatedAt   time.Time
	Status      EmailStatus
	Attempts    int
	LastError   string
}

// Attachment represents an email attachment
type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

// Priority represents email priority
type Priority int

const (
	PriorityLow Priority = iota
	PriorityNormal
	PriorityHigh
)


type EmailStatus string

const (
	StatusPending   EmailStatus = "pending"
	StatusSending   EmailStatus = "sending"
	StatusSent      EmailStatus = "sent"
	StatusFailed    EmailStatus = "failed"
	StatusRetrying  EmailStatus = "retrying"
)


// Validate checks if the email is valid
func (e *Email) Validate() error {
	if e.From == "" {
		return fmt.Errorf("from address is required")
	}
	
	if !isValidEmail(e.From) {
		return fmt.Errorf("invalid from address: %s", e.From)
	}
	
	if len(e.To) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}
	
	for _, addr := range e.To {
		if !isValidEmail(addr) {
			return fmt.Errorf("invalid to address: %s", addr)
		}
	}
	
	if e.Subject == "" {
		return fmt.Errorf("subject is required")
	}
	
	if e.TextBody == "" && e.HTMLBody == "" {
		return fmt.Errorf("at least one body (text or HTML) is required")
	}
	
	return nil
}

// isValidEmail validates email address format
func isValidEmail(email string) bool {
	// RFC 5322 simplified regex
	pattern := `^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(pattern)
	return re.MatchString(email)
}

// EmailBuilder provides a fluent interface for building emails
type EmailBuilder struct {
	email *Email
}


// NewEmailBuilder creates a new EmailBuilder
func NewEmailBuilder() *EmailBuilder {
	return &EmailBuilder{
		email: &Email{
			Headers:   make(map[string]string),
			Priority:  PriorityNormal,
			Status:    StatusPending,
			CreatedAt: time.Now(),
		},
	}
}

func (b *EmailBuilder) From(from string) *EmailBuilder {
	b.email.From = from
	return b
}

func (b *EmailBuilder) To(to ...string) *EmailBuilder {
	b.email.To = append(b.email.To, to...)
	return b
}

func (b *EmailBuilder) Cc(cc ...string) *EmailBuilder {
	b.email.Cc = append(b.email.Cc, cc...)
	return b
}

func (b *EmailBuilder) Bcc(bcc ...string) *EmailBuilder {
	b.email.Bcc = append(b.email.Bcc, bcc...)
	return b
}

func (b *EmailBuilder) Subject(subject string) *EmailBuilder {
	b.email.Subject = subject
	return b
}

func (b *EmailBuilder) TextBody(body string) *EmailBuilder {
	b.email.TextBody = body
	return b
}

func (b *EmailBuilder) HTMLBody(body string) *EmailBuilder {
	b.email.HTMLBody = body
	return b
}

func (b *EmailBuilder) Attach(filename, contentType string, data []byte) *EmailBuilder {
	b.email.Attachments = append(b.email.Attachments, Attachment{
		Filename:    filename,
		ContentType: contentType,
		Data:        data,
	})
	return b
}

func (b *EmailBuilder) Priority(priority Priority) *EmailBuilder {
	b.email.Priority = priority
	return b
}

func (b *EmailBuilder) Build() (*Email, error) {
	if err := b.email.Validate(); err != nil {
		return nil, err
	}
	return b.email, nil
}

func (b *EmailBuilder) Header(key, value string) *EmailBuilder {
	b.email.Headers[key] = value
	return b
}