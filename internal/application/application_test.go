package application_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"yandexlyceum/yandex_lyceum_go/internal/application"
)

func TestErrorMiddleware(t *testing.T) {
	testCases := []struct {
		name           string
		expression     string
		expectedStatus int
	}{
		{"Ok", "2+2/2", http.StatusOK},
		{"Internal error", "2/0", http.StatusInternalServerError},
		{"Invalid expression", "2/(", http.StatusUnprocessableEntity},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requestBody := map[string]string{"expression": tc.expression}
			requestBodyJSON, _ := json.Marshal(requestBody)
			handler := application.ErrorMiddleware(application.CalculateHandler)
			req := httptest.NewRequest("POST", "/api/v1/calculate", bytes.NewBuffer(requestBodyJSON))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Errorf("Expected status code %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}
