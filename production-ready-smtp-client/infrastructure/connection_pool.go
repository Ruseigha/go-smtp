package infrastructure

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"sync"
	"time"
)

type ConnectionPool struct {
	config      *SMTPConfig
	connections chan *smtp.Client
	mu          sync.Mutex
	closed      bool
}

func NewConnectionPool(config *SMTPConfig, size int) (*ConnectionPool, error) {
	pool := &ConnectionPool{
		config:      config,
		connections: make(chan *smtp.Client, size),
	}
	
	// Pre-populate pool
	for i := 0; i < size; i++ {
		conn, err := pool.createConnection()
		if err != nil {
			pool.Close()
			return nil, fmt.Errorf("failed to create connection: %w", err)
		}
		pool.connections <- conn
	}
	
	return pool, nil
}


func (p *ConnectionPool) createConnection() (*smtp.Client, error) {
	addr := p.config.Host + ":" + p.config.Port
	
	// For port 465 (implicit TLS)
	if p.config.Port == "465" {
		tlsConfig := &tls.Config{
			ServerName: p.config.Host,
		}
		
		conn, err := tls.Dial("tcp", addr, tlsConfig)
		if err != nil {
			return nil, fmt.Errorf("tls dial failed: %w", err)
		}
		
		client, err := smtp.NewClient(conn, p.config.Host)
		if err != nil {
			return nil, fmt.Errorf("smtp client failed: %w", err)
		}
		
		// Authenticate
		auth := smtp.PlainAuth("", p.config.Username, p.config.Password, p.config.Host)
		if err := client.Auth(auth); err != nil {
			client.Close()
			return nil, fmt.Errorf("auth failed: %w", err)
		}
		
		return client, nil
	}
	
	// For port 587 (STARTTLS)
	client, err := smtp.Dial(addr)
	if err != nil {
		return nil, fmt.Errorf("dial failed: %w", err)
	}
	
	if err := client.Hello("localhost"); err != nil {
		client.Close()
		return nil, fmt.Errorf("hello failed: %w", err)
	}
	
	tlsConfig := &tls.Config{
		ServerName: p.config.Host,
	}
	
	if err := client.StartTLS(tlsConfig); err != nil {
		client.Close()
		return nil, fmt.Errorf("starttls failed: %w", err)
	}
	
	auth := smtp.PlainAuth("", p.config.Username, p.config.Password, p.config.Host)
	if err := client.Auth(auth); err != nil {
		client.Close()
		return nil, fmt.Errorf("auth failed: %w", err)
	}
	
	return client, nil
}

func (p *ConnectionPool) Get(ctx context.Context) (*smtp.Client, error) {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, fmt.Errorf("pool is closed")
	}
	p.mu.Unlock()
	
	select {
	case conn := <-p.connections:
		// Test if connection is still alive
		if err := conn.Noop(); err != nil {
			// Connection is dead, create new one
			conn.Close()
			newConn, err := p.createConnection()
			if err != nil {
				return nil, fmt.Errorf("failed to recreate connection: %w", err)
			}
			return newConn, nil
		}
		return conn, nil
		
	case <-ctx.Done():
		return nil, ctx.Err()
		
	case <-time.After(30 * time.Second):
		return nil, fmt.Errorf("timeout waiting for connection")
	}
}

func (p *ConnectionPool) Put(conn *smtp.Client) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.closed {
		conn.Close()
		return nil
	}
	
	select {
	case p.connections <- conn:
		return nil
	default:
		// Pool is full, close connection
		conn.Close()
		return nil
	}
}

func (p *ConnectionPool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	
	if p.closed {
		return nil
	}
	
	p.closed = true
	close(p.connections)
	
	// Close all connections
	for conn := range p.connections {
		conn.Quit()
		conn.Close()
	}
	
	return nil
}