package services

import "log"

// FakeEmailSender — заглушка для разработки, не отправляет реальные письма.
type FakeEmailSender struct{}

func (s *FakeEmailSender) SendVerificationCode(email, code string) error {
	log.Printf("[FAKE EMAIL] To: %s, Code: %s", email, code)
	return nil
}
