package auth

import "testing"

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"валидный email", "test@mail.ru", true},
		{"валидный с точкой", "test.name@mail.ru", true},
		{"пустой email", "", false},
		{"без @", "testmail.ru", false},
		{"без домена", "test@", false},
		{"без точки в домене", "test@mailru", false},
		{"пробел в начале", "  test@mail.ru", true}, // TrimSpace обрезает
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ValidateEmail(tt.email)
			if got != tt.want {
				t.Errorf("ValidateEmail(%q) = %v, want %v", tt.email, got, tt.want)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"валидный пароль", "Pass123!", false},
		{"слишком короткий", "Pa1!", true},
		{"слишком длинный", string(make([]byte, 101)), true},
		{"только буквы", "Password", true},
		{"только цифры", "12345678", true},
		{"без спецсимвола", "Password1", true},
		{"с пробелом в начале", "  Pass1!23", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword(%q) error = %v, wantErr %v", tt.password, err, tt.wantErr)
			}
		})
	}
}

func TestIsPasswordStrong(t *testing.T) {
	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{"буква + цифра + спецсимвол", "Pass1!", true},
		{"только буквы", "Password", false},
		{"только цифры", "12345678", false},
		{"буква + цифра без спецсимвола", "Password1", false},
		{"кириллица + цифра + спецсимвол", "Пароль1!", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPasswordStrong(tt.password)
			if got != tt.want {
				t.Errorf("isPasswordStrong(%q) = %v, want %v", tt.password, got, tt.want)
			}
		})
	}
}
