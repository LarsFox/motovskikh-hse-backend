package emailer

import (
	"context"
	"log"
)

// Client — заглушка для разработки, не отправляет реальные письма.
type Client struct{}

func (c *Client) SendEmail(_ context.Context, to, subj, msg string) error {
	log.Printf("[FAKE EMAIL] To: %s, Subj: %s MSG: %s", to, subj, msg)
	return nil
}
