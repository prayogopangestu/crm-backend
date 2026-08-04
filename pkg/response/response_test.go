package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSuccess(t *testing.T) {
	w := httptest.NewRecorder()
	Success(w, http.StatusOK, "ok", map[string]int{"a": 1})

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var body APIResponse
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !body.Success || body.Message != "ok" {
		t.Fatalf("unexpected response: %+v", body)
	}
	data, ok := body.Data.(map[string]any)
	if !ok || data["a"] != float64(1) {
		t.Fatalf("Data not propagated: %+v", body.Data)
	}
	if body.Error != nil {
		t.Fatalf("Error should be nil on Success: %+v", body.Error)
	}
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	Error(w, http.StatusBadRequest, "bad")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}

	var body APIResponse
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.Success {
		t.Fatalf("Status should be false on Error")
	}
	if body.Message != "bad" {
		t.Fatalf("Message = %q, want %q", body.Message, "bad")
	}
	if body.Data != nil {
		t.Fatalf("Data should be nil on Error: %+v", body.Data)
	}
}
