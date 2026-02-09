package main

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func readResponse(reader *bufio.Reader) string {
	response, _ := reader.ReadString('\n')
	fmt.Printf("<<< %s", response)
	return response
}


func sendCommand(conn net.Conn, reader *bufio.Reader, command string) string {
	fmt.Fprintf(conn, "%s\r\n", command)
	fmt.Printf(">>> %s\n", command)
	return readResponse(reader)
}


func main()  {
	_ = godotenv.Load()
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	email := os.Getenv("SMTP_FROM")
	password := os.Getenv("SMTP_PASSWORD")
	to := "s.rufus.cse2023075@student.oauife.edu.ng"


	// Connect
	fmt.Println("🔌 Connecting to SMTP server...")
	conn, err := net.Dial("tcp", smtpHost+":"+smtpPort)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// Read greeting
	readResponse(reader)

	// EHLO
	sendCommand(conn, reader, "EHLO localhost")


	// Read capabilities
	for {
		line, _ := reader.ReadString('\n')
		fmt.Printf("<<< %s", line)
		if !strings.HasPrefix(line, "250-") {
			break
		}
	}

	// STARTTLS - Required by Gmail before authentication
	response := sendCommand(conn, reader, "STARTTLS")
	if !strings.HasPrefix(response, "220") {
		log.Fatalf("STARTTLS failed: %s", response)
	}

	// Upgrade connection to TLS (client-side)
	tlsConn := tls.Client(conn, &tls.Config{
		ServerName:         smtpHost,
		InsecureSkipVerify: false,
	})
	defer tlsConn.Close()

	// Create new reader for TLS connection
	reader = bufio.NewReader(tlsConn)

	// AUTH PLAIN
	// Format: \0username\0password (base64 encoded)
	auth := fmt.Sprintf("\x00%s\x00%s", email, password)
	authEncoded := base64.StdEncoding.EncodeToString([]byte(auth))
	
	fmt.Println(">>> AUTH PLAIN [base64 credentials]")
	fmt.Fprintf(tlsConn, "AUTH PLAIN %s\r\n", authEncoded)
	response = readResponse(reader)

	if !strings.HasPrefix(response, "235") {
		log.Fatalf("Authentication failed: %s", response)
	}

	fmt.Println("🔐 Authentication successful!")

	// MAIL FROM
	sendCommand(tlsConn, reader, fmt.Sprintf("MAIL FROM:<%s>", email))

	// RCPT TO
	sendCommand(tlsConn, reader, fmt.Sprintf("RCPT TO:<%s>", to))

	// DATA
	sendCommand(tlsConn, reader, "DATA")

	// Email content
	emailContent := fmt.Sprintf(`From: %s
To: %s
Subject: Authenticated Manual SMTP

This email was sent using:
1. Manual TCP connection
2. Manual SMTP commands
3. Manual authentication

We're SMTP hackers now! 🕵️
.`, email, to)

	fmt.Fprintf(tlsConn, "%s\r\n", emailContent)
	fmt.Println(">>> [Email content]")
	fmt.Println(">>> .")
	readResponse(reader)

	// QUIT
	sendCommand(tlsConn, reader, "QUIT")
	fmt.Println("\n✅ Authenticated email sent successfully!")
}