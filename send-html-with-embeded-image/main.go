package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type HTMLEmail struct {
	From    string
	To      []string
	Subject string
	HTML    string
	Images  map[string][]byte // cid -> image data
}

func (e *HTMLEmail) Build() []byte {
	var buf bytes.Buffer

	boundary := fmt.Sprintf("boundary_%d", time.Now().UnixNano())

	// Headers
	buf.WriteString(fmt.Sprintf("From: %s\r\n", e.From))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(e.To, ", ")))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", e.Subject))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/related; boundary=\"%s\"\r\n", boundary))
	buf.WriteString("\r\n")

	// HTML part
	buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(e.HTML)
	buf.WriteString("\r\n\r\n")

	// Embedded images
	for cid, imgData := range e.Images {
		buf.WriteString(fmt.Sprintf("--%s\r\n", boundary))
		buf.WriteString("Content-Type: image/png\r\n")
		buf.WriteString("Content-Transfer-Encoding: base64\r\n")
		buf.WriteString(fmt.Sprintf("Content-ID: <%s>\r\n", cid))
		buf.WriteString("Content-Disposition: inline\r\n")
		buf.WriteString("\r\n")

		// Encode image to base64
		encoded := base64.StdEncoding.EncodeToString(imgData)

		// Split into 76-character lines
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

	// End boundary
	buf.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	return buf.Bytes()
}

func main() {
	_ = godotenv.Load("../.env")

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	from := os.Getenv("SMTP_FROM")
	password := os.Getenv("SMTP_PASSWORD")


	// Read logo image (create a simple 1x1 PNG if file doesn't exist)
	logoData, err := os.ReadFile("logo.png")
	if err != nil {
		// Create a minimal 1x1 transparent PNG as fallback
		logoData = []byte{
			0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
			0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
			0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
			0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
			0x89, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41,
			0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
			0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00,
			0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE,
			0x42, 0x60, 0x82,
		}
	}


	// Create email
	email := HTMLEmail{
		From:    from,
		To:      []string{"s.rufus.cse2023075@student.oauife.edu.ng"},
		Subject: "HTML Email with Embedded Logo",
		HTML: `<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; max-width: 600px; margin: 0 auto; }
        .header { background: #4A90E2; color: white; padding: 30px; text-align: center; }
        .content { padding: 30px; background: #f5f5f5; }
        .logo { max-width: 100px; }
    </style>
</head>
<body>
    <div class="header">
        <img src="cid:logo" class="logo" alt="Logo" />
        <h1>Welcome to Our Service!</h1>
    </div>
    <div class="content">
        <p>Hello!</p>
        <p>This email demonstrates how to embed images directly in HTML emails.</p>
        <p>The logo above is embedded using <code>cid:logo</code> reference.</p>
        <p><strong>Best regards,</strong><br>The Team</p>
    </div>
</body>
</html>`,
		Images: map[string][]byte{
			"logo": logoData,
		},
	}

	// Build message
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

	fmt.Println("✅ HTML email with embedded image sent!")

}