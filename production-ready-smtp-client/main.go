package main

import (
	"context"
	"fmt"
	"go-smtp/production-ready-smtp-client/application"
	"go-smtp/production-ready-smtp-client/config"
	"go-smtp/production-ready-smtp-client/domain"
	"go-smtp/production-ready-smtp-client/infrastructure"
	"log"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load("../.env"); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create SMTP client
	smtpConfig := &infrastructure.SMTPConfig{
		Host:     cfg.SMTP.Host,
		Port:     cfg.SMTP.Port,
		Username: cfg.SMTP.Username,
		Password: cfg.SMTP.Password,
		PoolSize: cfg.SMTP.PoolSize,
	}

	smtpClient, err := infrastructure.NewSMTPClient(smtpConfig)
	if err != nil {
		log.Fatalf("Failed to create SMTP client: %v", err)
	}
	defer smtpClient.Close()

	// Create email service
	emailService := application.NewEmailService(smtpClient)
	defer emailService.Close()

	fmt.Println("🚀 Production-Ready SMTP Client Started")
	fmt.Printf("📧 SMTP Server: %s:%s\n", cfg.SMTP.Host, cfg.SMTP.Port)
	fmt.Printf("🔄 Connection Pool Size: %d\n\n", cfg.SMTP.PoolSize)

	// Example 1: Send simple email
	fmt.Println("📤 Example 1: Sending simple email...")
	if err := sendSimpleEmail(emailService, cfg.SMTP.Username); err != nil {
		log.Printf("❌ Failed: %v", err)
	}

	// Example 2: Send email with attachments
	fmt.Println("\n📤 Example 2: Sending email with attachments...")
	if err := sendEmailWithAttachments(emailService, cfg.SMTP.Username); err != nil {
		log.Printf("❌ Failed: %v", err)
	}

	// Example 3: Send bulk emails
	fmt.Println("\n📤 Example 3: Sending bulk emails...")
	if err := sendBulkEmails(emailService, cfg.SMTP.Username); err != nil {
		log.Printf("❌ Failed: %v", err)
	}

	fmt.Println("\n✅ All examples completed!")
}


func sendSimpleEmail(service *application.EmailService, from string) error {
	email, err := domain.NewEmailBuilder().
		From(from).
		To("recipient@example.com").
		Subject("Production SMTP Client Test").
		TextBody("This email was sent using a production-ready SMTP client with:\n- Connection pooling\n- Retry logic\n- Clean architecture").
		HTMLBody(`<html>
<body>
	<h1>Production SMTP Client Test</h1>
	<p>This email was sent using a production-ready SMTP client with:</p>
	<ul>
		<li>Connection pooling</li>
		<li>Retry logic with exponential backoff</li>
		<li>Clean architecture</li>
	</ul>
</body>
</html>`).
		Build()
	
	if err != nil {
		return err
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	return service.SendEmail(ctx, email)
}

func sendEmailWithAttachments(service *application.EmailService, from string) error {
	// Create sample attachment
	attachmentData := []byte("This is a sample text file.\nIt was attached to an email.\n")
	
	email, err := domain.NewEmailBuilder().
		From(from).
		To("s.rufus.cse2023075@student.oauife.edu.ng").
		Subject("📎 Email with Attachment").
		TextBody("Please find the attached file.").
		HTMLBody(`<html>
<body>
	<h2>📎 Email with Attachment</h2>
	<p>Please find the attached file.</p>
</body>
</html>`).
		Attach("sample.txt", "text/plain", attachmentData).
		Priority(domain.PriorityHigh).
		Build()
	
	if err != nil {
		return err
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	
	return service.SendEmail(ctx, email)
}

func sendBulkEmails(service *application.EmailService, from string) error {
	var emails []*domain.Email
	
	for i := 1; i <= 3; i++ {
		email, err := domain.NewEmailBuilder().
			From(from).
			To("s.rufus.cse2023075@student.oauife.edu.ng").
			Subject(fmt.Sprintf("Bulk Email #%d", i)).
			TextBody(fmt.Sprintf("This is bulk email number %d", i)).
			HTMLBody(fmt.Sprintf(`<html><body><h2>Bulk Email #%d</h2></body></html>`, i)).
			Build()
		
		if err != nil {
			return err
		}
		
		emails = append(emails, email)
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	
	return service.SendBulkEmails(ctx, emails)
}