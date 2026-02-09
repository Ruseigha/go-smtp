package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"mime"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Email struct {
	From        string
	To          []string
	Subject     string
	TextBody    string
	HTMLBody    string
	Attachments []Attachment
}

type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

func generateBoundary() string {
	return fmt.Sprintf("boundary_%d", time.Now().UnixNano())
}

func encodeAttachment(data []byte) string {
	encoded := base64.StdEncoding.EncodeToString(data)

	// Split into 76-character lines
	var result strings.Builder
	for i := 0; i < len(encoded); i += 76 {
		end := i + 76
		if end > len(encoded) {
			end = len(encoded)
		}
		result.WriteString(encoded[i:end])
		result.WriteString("\r\n")
	}

	return result.String()
}

func (e *Email) Build() []byte {
	var buffer bytes.Buffer

	// Headers
	buffer.WriteString(fmt.Sprintf("From: %s\r\n", e.From))
	buffer.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(e.To, ", ")))
	buffer.WriteString(fmt.Sprintf("Subject: %s\r\n", mime.BEncoding.Encode("UTF-8", e.Subject)))
	buffer.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	buffer.WriteString(fmt.Sprintf("Message-ID: <%d@example.com>\r\n", time.Now().UnixNano()))
	buffer.WriteString("MIME-Version: 1.0\r\n")

	// Main boundary for multipart/mixed (attachments)
	outerBoundary := generateBoundary()
	buffer.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", outerBoundary))
	buffer.WriteString("\r\n")

	// First part: multipart/alternative (text + HTML)
	buffer.WriteString(fmt.Sprintf("--%s\r\n", outerBoundary))

	innerBoundary := generateBoundary()
	buffer.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", innerBoundary))
	buffer.WriteString("\r\n")

	// Plain text version
	buffer.WriteString(fmt.Sprintf("--%s\r\n", innerBoundary))
	buffer.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	buffer.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	buffer.WriteString("\r\n")
	buffer.WriteString(e.TextBody)
	buffer.WriteString("\r\n\r\n")

	// HTML version
	buffer.WriteString(fmt.Sprintf("--%s\r\n", innerBoundary))
	buffer.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	buffer.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	buffer.WriteString("\r\n")
	buffer.WriteString(e.HTMLBody)
	buffer.WriteString("\r\n\r\n")

	// End inner boundary
	buffer.WriteString(fmt.Sprintf("--%s--\r\n", innerBoundary))
	buffer.WriteString("\r\n")

	// Attachments
	for _, att := range e.Attachments {
		buffer.WriteString(fmt.Sprintf("--%s\r\n", outerBoundary))
		buffer.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"\r\n", att.ContentType, att.Filename))
		buffer.WriteString("Content-Transfer-Encoding: base64\r\n")
		buffer.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", att.Filename))
		buffer.WriteString("\r\n")
		buffer.WriteString(encodeAttachment(att.Data))
		buffer.WriteString("\r\n")
	}

	// End outer boundary
	buffer.WriteString(fmt.Sprintf("--%s--\r\n", outerBoundary))

	return buffer.Bytes()
}

func main() {
	_ = godotenv.Load("../.env")

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	from := os.Getenv("SMTP_FROM")
	password := os.Getenv("SMTP_PASSWORD")

	// Read attachment
	pdfData, err := os.ReadFile("document.pdf")
	if err != nil {
		log.Printf("Warning: Could not read PDF: %v", err)
		pdfData = []byte("Sample PDF content")
	}

	// Create email
	email := Email{
		From:    from,
		To:      []string{"s.rufus.cse2023075@student.oauife.edu.ng"},
		Subject: "📧 Email with Attachments - MIME Complete Example",
		TextBody: `Hello!

This is a professional email with attachments.

Contents:
- Plain text version (you're reading this)
- HTML version (for modern email clients)
- PDF attachment
- Image attachment

Best regards,
Your Go SMTP System`,
		HTMLBody: `<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); 
                  color: white; padding: 30px; border-radius: 10px; text-align: center; }
        .content { background: #f9f9f9; padding: 20px; margin-top: 20px; border-radius: 5px; }
        .footer { margin-top: 30px; font-size: 14px; color: #666; text-align: center; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>📧 Email with Attachments</h1>
            <p>MIME Complete Example</p>
        </div>
        
        <div class="content">
            <p>Hello!</p>
            
            <p>This is a professional email with attachments.</p>
            
            <h3>Contents:</h3>
            <ul>
                <li>✅ Plain text version</li>
                <li>✅ HTML version (you're reading this)</li>
                <li>✅ PDF attachment</li>
                <li>✅ Image attachment</li>
            </ul>
            
            <p><strong>Best regards,</strong><br>Your Go SMTP System</p>
        </div>
        
        <div class="footer">
            <p>Sent using Go SMTP with full MIME support</p>
        </div>
    </div>
</body>
</html>`,
		Attachments: []Attachment{
			{
				Filename:    "document.pdf",
				ContentType: "application/pdf",
				Data:        pdfData,
			},
		},
	}

	// Build email
	message := email.Build()

	// Send
	auth := smtp.PlainAuth("", from, password, smtpHost)
	err = smtp.SendMail(
		smtpHost+":"+smtpPort,
		auth,
		from,
		email.To,
		message,
	)

	if err != nil {
		log.Fatalf("Failed to send: %v", err)
	}

	fmt.Println("✅ Email with attachments sent successfully!")
	fmt.Printf("   Attachments: %d\n", len(email.Attachments))

}
