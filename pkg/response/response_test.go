package response

import "testing"

func TestSuccess(t *testing.T) {
	got := Success("ok", map[string]int{"a": 1})
	if !got.Status || got.Message != "ok" {
		t.Fatalf("unexpected response: %+v", got)
	}
	data, ok := got.Data.(map[string]int)
	if !ok || data["a"] != 1 {
		t.Fatalf("Data not propagated: %+v", got.Data)
	}
	if got.Errors != nil {
		t.Fatalf("Errors should be nil on Success: %+v", got.Errors)
	}
}

func TestError(t *testing.T) {
	got := Error("bad", map[string]string{"field": "required"})
	if got.Status {
		t.Fatalf("Status should be false on Error")
	}
	if got.Message != "bad" {
		t.Fatalf("Message = %q, want %q", got.Message, "bad")
	}
	if got.Data != nil {
		t.Fatalf("Data should be nil on Error: %+v", got.Data)
	}
}
