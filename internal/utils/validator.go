package utils

import (
	"fmt"
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
var loginRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)

//корректность email
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if email == "" {
		return fmt.Errorf("Пустота не может быть email адресом")
	}
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("Некорректный формат email")
	}
	return nil
}

//корректность логина
func ValidateLogin(login string) error {
	login = strings.TrimSpace(login)
	if len(login) < 3 || len(login) > 20 {
		return fmt.Errorf("Логин должен быть от 3 до 20 символов")
	}
	if !loginRegex.MatchString(login) {
		return fmt.Errorf("Логин может содержать только буквы, цифры и нижнее подчеркивание")
	}
	return nil
}

//надежность пароля
func ValidatePassword(password string) error {
	if len(password) < 6 {
		return fmt.Errorf("Пароль должен быть минимум 6 символов")
	}
	if len(password) > 100 {
		return fmt.Errorf("Не мучайте свою память! Пароль слишком длинный")
	}
	return nil
}
