package storage

import "context"

type Client struct {
	endpoint    string
	accessKey   string
	secretKey   string
	bucket      string
	useSSL      bool
}

type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

func NewClient(cfg Config) *Client {
	return &Client{
		endpoint:  cfg.Endpoint,
		accessKey: cfg.AccessKey,
		secretKey: cfg.SecretKey,
		bucket:    cfg.Bucket,
		useSSL:    cfg.UseSSL,
	}
}

func (c *Client) Upload(ctx context.Context, key string, data []byte, contentType string) error {
	return nil
}

func (c *Client) GetURL(key string) string {
	return ""
}

func (c *Client) Delete(ctx context.Context, key string) error {
	return nil
}
