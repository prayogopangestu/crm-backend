package email

import (
	"fmt"
	"net/smtp"
)

type Client struct {
	host     string
	port     string
	username string
	password string
}

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
}

func NewClient(cfg Config) *Client {
	return &Client{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.Username,
		password: cfg.Password,
	}
}

func (c *Client) Send(to, subject, body string) error {
	addr := c.host + ":" + c.port
	auth := smtp.PlainAuth("", c.username, c.password, c.host)
	msg := []byte("Subject: " + subject + "\r\n" +
		"From: " + c.username + "\r\n" +
		"To: " + to + "\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n\r\n" +
		body)
	return smtp.SendMail(addr, auth, c.username, []string{to}, msg)
}

func (c *Client) Ping() error {
	if c.host == "" || c.port == "" {
		return fmt.Errorf("email client not configured")
	}
	return nil
}
