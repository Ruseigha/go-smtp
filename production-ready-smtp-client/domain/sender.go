package domain

import "context"

// EmailSender is the interface for sending emails
type EmailSender interface {
	Send(ctx context.Context, email *Email) error
	SendBulk(ctx context.Context, emails []*Email) error
	Close() error
}

// EmailRepository is the interface for storing emails
type EmailRepository interface {
	Save(ctx context.Context, email *Email) error
	GetByID(ctx context.Context, id string) (*Email, error)
	GetPending(ctx context.Context, limit int) ([]*Email, error)
	UpdateStatus(ctx context.Context, id string, status EmailStatus) error
}