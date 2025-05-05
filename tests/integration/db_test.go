package tests

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"yandexlyceum/internal/application"
	"yandexlyceum/internal/database"

	"github.com/golang-jwt/jwt/v5"
)

func TestCalculateHandlerWithDB(t *testing.T) {
	dbPath := "test_integration.db"
	_ = os.Remove(dbPath)

	db, err := database.InitDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to init DB: %v", err)
	}
	defer func() {
		db.Close()
		_ = os.Remove(dbPath)
	}()

	o := application.NewOrchestrator()
	o.Db = db
	o.Config = application.ConfigFromEnv()

	_, err = db.Exec("INSERT INTO users(login, password) VALUES(?, ?)",
		"testuser", "hashedpassword")
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	reqBody := `{"expression": "2+2"}`
	req := httptest.NewRequest("POST", "/api/v1/calculate", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")

	ctx := context.WithValue(req.Context(), application.UserContextKey, jwt.MapClaims{"user_id": float64(1)})
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handler := http.HandlerFunc(o.CalculateHandler)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM expressions WHERE user_id = 1").Scan(&count)
	if err != nil {
		t.Fatalf("DB query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 expression in DB, got %d", count)
	}
}
