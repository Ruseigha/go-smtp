package main

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type SMTPConfig struct {
	Host     string
	Port     string
	From     string
	Password string
}

func loadSMTPConfig() (*SMTPConfig, error) {
	// Try to load .env file (ignore error if it doesn't exist)
	_ = godotenv.Load()

	config := &SMTPConfig{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     os.Getenv("SMTP_PORT"),
		From:     os.Getenv("SMTP_FROM"),
		Password: os.Getenv("SMTP_PASSWORD"),
	}

	// Validate configuration
	missing := []string{}
	if config.Host == "" {
		missing = append(missing, "SMTP_HOST")
	}
	if config.Port == "" {
		missing = append(missing, "SMTP_PORT")
	}
	if config.From == "" {
		missing = append(missing, "SMTP_FROM")
	}
	if config.Password == "" {
		missing = append(missing, "SMTP_PASSWORD")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("missing environment variables: %s", strings.Join(missing, ", "))
	}

	return config, nil
}

func sendEmail(config *SMTPConfig, to []string, htmlBody string) error {
	subject := "Welcome to YourApp!"
	// Build email message
	message := []byte(
		"From: " + config.From + "\r\n" +
			"To: " + to[0] + "\r\n" +
			"Subject: " + subject + "\r\n" +
			"MIME-Version: 1.0\r\n" +
			"Content-Type: text/html; charset=UTF-8\r\n" +
			"\r\n" +
			htmlBody,
	)

	// Set up authentication
	auth := smtp.PlainAuth("", config.From, config.Password, config.Host)

	// Send email
	err := smtp.SendMail(config.Host+":"+config.Port, auth, config.From, to, message)

	return err
}

func main() {
	// Load SMTP configuration
	config, err := loadSMTPConfig()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	fmt.Println("📧 SMTP Configuration loaded successfully")
	fmt.Printf("   Host: %s:%s\n", config.Host, config.Port)
	fmt.Printf("   From: %s\n", config.From)
	fmt.Println()

	// Send test email
	to := []string{"s.rufus.cse2023075@student.oauife.edu.ng"} // Replace with your recipient
	subject := "Test Email from Go SMTP Server"
	htmlbody := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Welcome!</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            line-height: 1.6;
            color: #333;
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .container {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            border-radius: 10px;
            padding: 40px;
        }
        .content {
            background-color: white;
            border-radius: 8px;
            padding: 40px;
        }
        h1 {
            color: #333;
            text-align: center;
            font-size: 32px;
            margin-top: 0;
        }
        .emoji {
            font-size: 48px;
            text-align: center;
            margin-bottom: 20px;
        }
        .feature-list {
            background-color: #f8f9fa;
            border-radius: 8px;
            padding: 20px;
            margin: 20px 0;
        }
        .feature-item {
            padding: 10px 0;
            border-bottom: 1px solid #e0e0e0;
        }
        .feature-item:last-child {
            border-bottom: none;
        }
        .button {
            display: inline-block;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            text-decoration: none;
            padding: 14px 40px;
            border-radius: 6px;
            margin: 20px 0;
            font-weight: 600;
        }
        .footer {
            margin-top: 30px;
            text-align: center;
            font-size: 14px;
            color: #666;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="content">
            <div class="emoji">🎉</div>
            <h1>Welcome to YourApp!</h1>
            
            <p>Hi there!</p>
            
            <p>Your email has been verified and your account is now active. We're thrilled to have you on board!</p>
            
            <div class="feature-list">
                <h3 style="margin-top: 0;">What's Next?</h3>
                
                <div class="feature-item">
                    <strong>🔐 Secure Your Account</strong><br>
                    Enable two-factor authentication for extra security
                </div>
                
                <div class="feature-item">
                    <strong>👤 Complete Your Profile</strong><br>
                    Add your information to get the most out of YourApp
                </div>
                
                <div class="feature-item">
                    <strong>🚀 Explore Features</strong><br>
                    Check out all the amazing things you can do
                </div>
                
                <div class="feature-item">
                    <strong>💬 Get Support</strong><br>
                    Need help? Our support team is here for you
                </div>
            </div>
            
            <div style="text-align: center;">
                <a href="https://yourapp.com/dashboard" class="button">Go to Dashboard</a>
            </div>
            
            <p style="margin-top: 30px; font-size: 14px; color: #666;">
                If you have any questions or need assistance, don't hesitate to reach out to us at 
                <a href="mailto:support@yourapp.com">support@yourapp.com</a>
            </p>
        </div>
        
        <div class="footer">
            <p style="color: white; margin-top: 20px;">
                Thanks for choosing YourApp!
            </p>
        </div>
    </div>
</body>
</html>`

	fmt.Println("📤 Sending email...")
	err = sendEmail(config, to, htmlbody)
	if err != nil {
		log.Fatalf("❌ Failed to send email: %v", err)
	}

	fmt.Println("✅ Email sent successfully!")
	fmt.Printf("   To: %s\n", strings.Join(to, ", "))
	fmt.Printf("   Subject: %s\n", subject)
}
