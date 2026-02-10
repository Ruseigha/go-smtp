package application

import (
	"context"
	"fmt"
	"go-smtp/production-ready-smtp-client/domain"
	"go-smtp/production-ready-smtp-client/pkg/retry"
	"log"
)

type EmailService struct {
	sender      domain.EmailSender
	retryConfig retry.Config
}

func NewEmailService(sender domain.EmailSender) *EmailService {
	return &EmailService{
		sender:      sender,
		retryConfig: retry.DefaultConfig(),
	}
}

func (s *EmailService) SendEmail(ctx context.Context, email *domain.Email) error {
	// Validate email
	if err := email.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	
	// Update status
	email.Status = domain.StatusSending
	
	// Send with retry logic
	err := retry.Do(ctx, s.retryConfig, func(ctx context.Context) error {
		email.Attempts++
		return s.sender.Send(ctx, email)
	})
	
	if err != nil {
		email.Status = domain.StatusFailed
		email.LastError = err.Error()
		log.Printf("Failed to send email to %v: %v", email.To, err)
		return fmt.Errorf("failed to send email: %w", err)
	}
	
	email.Status = domain.StatusSent
	log.Printf("Successfully sent email to %v (attempts: %d)", email.To, email.Attempts)
	return nil
}


func (s *EmailService) SendBulkEmails(ctx context.Context, emails []*domain.Email) error {
	successCount := 0
	failCount := 0
	
	for _, email := range emails {
		if err := s.SendEmail(ctx, email); err != nil {
			failCount++
			log.Printf("Failed to send email: %v", err)
		} else {
			successCount++
		}
	}
	
	log.Printf("Bulk send completed: %d succeeded, %d failed", successCount, failCount)
	
	if failCount > 0 {
		return fmt.Errorf("failed to send %d out of %d emails", failCount, len(emails))
	}
	
	return nil
}

func (s *EmailService) SendWithTemplate(ctx context.Context, templateName string, data interface{}, email *domain.Email) error {
	// TODO: Implement template rendering
	// This would use the TemplateEngine to render HTML
	return s.SendEmail(ctx, email)
}

func (s *EmailService) Close() error {
	return s.sender.Close()
}