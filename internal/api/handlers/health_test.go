package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandlerGetHealth(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := NewHealthHandler()

	handler.GetHealth(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("GetHealth returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}
