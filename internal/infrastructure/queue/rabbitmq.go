package queue

import "context"

type Client struct {
	url string
}

type Config struct {
	URL string
}

func NewClient(cfg Config) *Client {
	return &Client{url: cfg.URL}
}

func (c *Client) Publish(ctx context.Context, queue string, message []byte) error {
	return nil
}

func (c *Client) Consume(ctx context.Context, queue string) (<-chan []byte, error) {
	ch := make(chan []byte)
	close(ch)
	return ch, nil
}

func (c *Client) Close() error {
	return nil
}
