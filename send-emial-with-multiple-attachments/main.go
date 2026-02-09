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

type EmailWithAttachments struct {
	From        string
	To          []string
	Subject     string
	TextBody    string
	HTMLBody    string
	Attachments []EmailAttachment
}

type EmailAttachment struct {
	Filename    string
	ContentType string
	Data        []byte
}

func (e *EmailWithAttachments) Build() []byte {
	var buf bytes.Buffer

	outerBoundary := fmt.Sprintf("outer_%d", time.Now().UnixNano())
	innerBoundary := fmt.Sprintf("inner_%d", time.Now().UnixNano()+1)

	// Headers
	buf.WriteString(fmt.Sprintf("From: %s\r\n", e.From))
	buf.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(e.To, ", ")))
	buf.WriteString(fmt.Sprintf("Subject: %s\r\n", mime.BEncoding.Encode("UTF-8", e.Subject)))
	buf.WriteString(fmt.Sprintf("Date: %s\r\n", time.Now().Format(time.RFC1123Z)))
	buf.WriteString("MIME-Version: 1.0\r\n")
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", outerBoundary))
	buf.WriteString("\r\n")

	// Email body (multipart/alternative: text + HTML)
	buf.WriteString(fmt.Sprintf("--%s\r\n", outerBoundary))
	buf.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", innerBoundary))
	buf.WriteString("\r\n")

	// Plain text version
	buf.WriteString(fmt.Sprintf("--%s\r\n", innerBoundary))
	buf.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(e.TextBody)
	buf.WriteString("\r\n\r\n")

	// HTML version
	buf.WriteString(fmt.Sprintf("--%s\r\n", innerBoundary))
	buf.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	buf.WriteString("\r\n")
	buf.WriteString(e.HTMLBody)
	buf.WriteString("\r\n\r\n")

	// End inner boundary
	buf.WriteString(fmt.Sprintf("--%s--\r\n", innerBoundary))
	buf.WriteString("\r\n")

	// Attachments
	for _, att := range e.Attachments {
		buf.WriteString(fmt.Sprintf("--%s\r\n", outerBoundary))
		buf.WriteString(fmt.Sprintf("Content-Type: %s; name=\"%s\"\r\n", att.ContentType, att.Filename))
		buf.WriteString("Content-Transfer-Encoding: base64\r\n")
		buf.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=\"%s\"\r\n", att.Filename))
		buf.WriteString("\r\n")

		// Encode to base64
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

	// End outer boundary
	buf.WriteString(fmt.Sprintf("--%s--\r\n", outerBoundary))

	return buf.Bytes()
}

func createSamplePDF() []byte {
	// Minimal valid PDF
	return []byte(`%PDF-1.4
1 0 obj

/Type /Catalog
/Pages 2 0 R
>>
endobj
2 0 obj

/Type /Pages
/Kids [3 0 R]
/Count 1
>>
endobj
3 0 obj

/Type /Page
/Parent 2 0 R
/MediaBox [0 0 612 792]
/Contents 4 0 R
/Resources 
/Font 
/F1 
/Type /Font
/Subtype /Type1
/BaseFont /Helvetica
>>
>>
>>
>>
endobj
4 0 obj

/Length 44
>>
stream
BT
/F1 12 Tf
100 700 Td
(Sample PDF) Tj
ET
endstream
endobj
xref
0 5
0000000000 65535 f 
0000000009 00000 n 
0000000058 00000 n 
0000000115 00000 n 
0000000317 00000 n 
trailer

/Size 5
/Root 1 0 R
>>
startxref
410
%%EOF`)
}

func main() {
	_ = godotenv.Load("../.env")

	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	from := os.Getenv("SMTP_FROM")
	password := os.Getenv("SMTP_PASSWORD")

	// Read files or use samples
	pdfData, err := os.ReadFile("invoice.pdf")
	if err != nil {
		pdfData = createSamplePDF()
	}

	textData := []byte("This is a sample text file.\nIt contains multiple lines.\n")

	// Create email
	email := EmailWithAttachments{
		From:    from,
		To:      []string{"s.rufus.cse2023075@student.oauife.edu.ng"},
		Subject: "📎 Invoice & Documents - Multiple Attachments",
		TextBody: `Dear Customer,

Please find attached:
- Invoice for January 2025
- Terms and conditions

Thank you for your business!

Best regards,
Billing Team`,
		HTMLBody: `<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #2C3E50; color: white; padding: 20px; border-radius: 5px; }
        .content { background: #ECF0F1; padding: 20px; margin-top: 20px; border-radius: 5px; }
        .attachments { margin-top: 20px; }
        .attachment { background: white; padding: 10px; margin: 5px 0; border-left: 4px solid #3498DB; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h2>📎 Invoice & Documents</h2>
        </div>
        <div class="content">
            <p>Dear Customer,</p>
            <p>Please find attached:</p>
            <div class="attachments">
                <div class="attachment">📄 Invoice for January 2025</div>
                <div class="attachment">📝 Terms and conditions</div>
            </div>
            <p><strong>Thank you for your business!</strong></p>
            <p>Best regards,<br><em>Billing Team</em></p>
        </div>
    </div>
</body>
</html>`,
		Attachments: []EmailAttachment{
			{
				Filename:    "invoice_january_2025.pdf",
				ContentType: "application/pdf",
				Data:        pdfData,
			},
			{
				Filename:    "terms.txt",
				ContentType: "text/plain",
				Data:        textData,
			},
		},
	}

	// Build and send
	message := email.Build()

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

	fmt.Println("✅ Email sent with attachments!")
	fmt.Printf("   Attachments: %d\n", len(email.Attachments))
	for _, att := range email.Attachments {
		fmt.Printf("   - %s (%d bytes)\n", att.Filename, len(att.Data))
	}
}
