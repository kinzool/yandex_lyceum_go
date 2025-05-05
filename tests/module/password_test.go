package tests

import (
	"testing"
	"yandexlyceum/internal/application"

	"golang.org/x/crypto/bcrypt"
)

func TestGeneratePassword(t *testing.T) {
	testCases := []struct {
		name     string
		password string
	}{
		{"Обычный пароль", "mySecurePassword123"},
		{"Пустой пароль", ""},
		{"Длинный пароль", "veryLongPassword" + string(make([]byte, 72))},
		{"Спецсимволы", "!@#$%^&*()"},
		{"Юникод", "парольΔtest"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hash, err := application.GeneratePassword(tc.password)
			if err != nil {
				t.Fatalf("GeneratePassword вернула ошибку: %v", err)
			}

			if hash == tc.password {
				t.Error("Хеш пароля совпадает с оригиналом")
			}

			err = bcrypt.CompareHashAndPassword([]byte(hash), []byte(tc.password))
			if err != nil {
				t.Errorf("bcrypt не смог верифицировать хеш: %v", err)
			}
		})
	}
}
