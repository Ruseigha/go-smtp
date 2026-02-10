package infrastructure

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"go-smtp/production-ready-smtp-client/domain"
	"mime"
	"net/smtp"
	"strings"
	"time"
)

type SMTPConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	PoolSize int
}

type SMTPClient struct {
	config *SMTPConfig
	pool   *ConnectionPool
}

func NewSMTPClient(config *SMTPConfig) (*SMTPClient, error) {
	pool, err := NewConnectionPool(config, config.PoolSize)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	return &SMTPClient{
		config: config,
		pool:   pool,
	}, nil
}


func (c *SMTPClient) Send(ctx context.Context, email *domain.Email) error {
	// Get connection from pool
	conn, err := c.pool.Get(ctx)
	if err != nil {
		return fmt.Errorf("failed to get connection: %w", err)
	}
	defer c.pool.Put(conn)
	
	// Build message
	message, err := c.buildMessage(email)
	if err != nil {
		return fmt.Errorf("failed to build message: %w", err)
	}
	
	// Send using connection
	if err := c.sendWithConnection(conn, email, message); err != nil {
		return fmt.Errorf("failed to send: %w", err)
	}
	
	return nil
}

func (c *SMTPClient) SendBulk(ctx context.Context, emails []*domain.Email) error {
	// Use goroutines for concurrent sending
	errChan := make(chan error, len(emails))
	
	for _, email := range emails {
		go func(e *domain.Email) {
			errChan <- c.Send(ctx, e)
		}(email)
	}
	
	// Collect errors
	var errors []error
	for i := 0; i < len(emails); i++ {
		if err := <-errChan; err != nil {
			errors = append(errors, err)
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("failed to send %d emails: %v", len(errors), errors[0])
	}
	
	return nil
}

func (c *SMTPClient) sendWithConnection(conn *smtp.Client, email *domain.Email, message []byte) error {
	// MAIL FROM
	if err := conn.Mail(email.From); err != nil {
		return fmt.Errorf("MAIL FROM failed: %w", err)
	}
	
	// RCPT TO
	allRecipients := append(append(email.To, email.Cc...), email.Bcc...)
	for _, addr := range allRecipients {
		if err := conn.Rcpt(addr); err != nil {
			return fmt.Errorf("RCPT TO %s failed: %w", addr, err)
		}
	}
	
	// DATA
	w, err := conn.Data()
	if err != nil {
		return fmt.Errorf("DATA failed: %w", err)
	}
	
	if _, err := w.Write(message); err != nil {
		return fmt.Errorf("write failed: %w", err)
	}
	
	if err := w.Close(); err != nil {
		return fmt.Errorf("close failed: %w", err)
	}
	
	// Reset for next email
	return conn.Reset()
}

func (c *SMTPClient) buildMessage(email *domain.Email) ([]byte, error) {
	var buf bytes.Buffer
	
	// Generate boundaries
	outerBoundary := fmt.Sprintf("outer_%d", time.Now().UnixNano())
	innerBoundary := fmt.Sprintf("inner_%d", time.Now().UnixNano()+1)
	
	// Headers
	buf.WriteString(fmt.Sprintf("From: %s\r\n", email.From))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(email.To, ", ")))
	
	if len(email.Cc) > 0 {
		buf.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(email.Cc, ", ")))
	}
	
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", mime.BEncoding.Encode("UTF-8", email.Subject)))
	buf.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	buf.WriteString(fmt.Sprintf("Message-ID: <%d@%s>\r\n", time.Now().UnixNano(), c.config.Host))
	
	// Custom headers
	for key, value := range email.Headers {
		buf.WriteString(fmt.Sprintf("%s: %s\r\n", key, value))
	}
	
	// Priority
	switch email.Priority {
	case domain.PriorityHigh:
		buf.WriteString("X-Priority: 1\r\n")
		buf.WriteString("Importance: high\r\n")
	case domain.PriorityLow:
		buf.WriteString("X-Priority: 5\r\n")
		buf.WriteString("Importance: low\r\n")
	}
	
	buf.WriteString("MIME-Version: 1.0\r\n")
	
	// If we have attachments, use multipart/mixed
	if len(email.Attachments) > 0 {
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", outerBoundary))
		buf.WriteString("\r\n")
		
		// Email body (multipart/alternative)
		buf.WriteString(fmt.Sprintf("--%s\r\n", outerBoundary))
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", innerBoundary))
		buf.WriteString("\r\n")
		
		// Text part
		if email.TextBody != "" {
			buf.WriteString(fmt.Sprintf("--%s\r\n", innerBoundary))
			buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
			buf.WriteString("\r\n")
			buf.WriteString(email.TextBody)
			buf.WriteString("\r\n\r\n")
		}
		
		// HTML part
		if email.HTMLBody != "" {
			buf.WriteString(fmt.Sprintf("--%s\r\n", innerBoundary))
			buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
			buf.WriteString("\r\n")
			buf.WriteString(email.HTMLBody)
			buf.WriteString("\r\n\r\n")
		}
		
		buf.WriteString(fmt.Sprintf("--%s--\r\n", innerBoundary))
		buf.WriteString("\r\n")
		
		// Attachments
		for _, att := range email.Attachments {
			buf.WriteString(fmt.Sprintf("--%s\r\n", outerBoundary))
			buf.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"\r\n", att.ContentType, att.Filename))
			buf.WriteString("Content-Transfer-Encoding: base64\r\n")
			buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", att.Filename))
			buf.WriteString("\r\n")
			
			encoded := base64.StdEncoding.EncodeToString(att.Data)
			for i := 0; i < len(encoded); i += 76 {
				end := i + 76
				if end > len(encoded) {
					end = len(encoded)
				}
				buf.WriteString(encoded[i:end])
				buf.WriteString("\r\n")
			}
			buf.WriteString("\r\n")
		}
		
		buf.WriteString(fmt.Sprintf("--%s--\r\n", outerBoundary))
	} else {
		// No attachments, use multipart/alternative
		buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", innerBoundary))
		buf.WriteString("\r\n")
		
		if email.TextBody != "" {
			buf.WriteString(fmt.Sprintf("--%s\r\n", innerBoundary))
			buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
			buf.WriteString("\r\n")
			buf.WriteString(email.TextBody)
			buf.WriteString("\r\n\r\n")
		}
		
		if email.HTMLBody != "" {
			buf.WriteString(fmt.Sprintf("--%s\r\n", innerBoundary))
			buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
			buf.WriteString("\r\n")
			buf.WriteString(email.HTMLBody)
			buf.WriteString("\r\n\r\n")
		}
		
		buf.WriteString(fmt.Sprintf("--%s--\r\n", innerBoundary))
	}
	
	return buf.Bytes(), nil
}

func (c *SMTPClient) Close() error {
	return c.pool.Close()
}

