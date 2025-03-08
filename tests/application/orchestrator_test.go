package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"yandexlyceum/internal/application"
)

type CalculateHandlerRequest struct {
	Expression string "'json:expression'"
}

func TestCalculateHandler(t *testing.T) {
	user := CalculateHandlerRequest{Expression: "2+2*2"}
	jsonBody, err := json.Marshal(user)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("POST", "/api/v1/calculate", bytes.NewBuffer(jsonBody))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(application.NewOrchestrator().CalculateHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestExpressionHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/api/v1/expressions", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(application.NewOrchestrator().ExpressionsHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
