package pagination

import "testing"

func TestPageRequestNormalized(t *testing.T) {
	tests := []struct {
		name         string
		input        PageRequest
		defaultLimit int
		maxLimit     int
		wantPage     int
		wantLimit    int
	}{
		{"empty defaults to 1/20", PageRequest{}, 20, 100, 1, 20},
		{"negative page becomes 1", PageRequest{Page: -3, Limit: 5}, 20, 100, 1, 5},
		{"limit clamped to max", PageRequest{Page: 2, Limit: 500}, 20, 100, 2, 100},
		{"valid values preserved", PageRequest{Page: 3, Limit: 50}, 20, 100, 3, 50},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input.Normalized(tc.defaultLimit, tc.maxLimit)
			if got.Page != tc.wantPage || got.Limit != tc.wantLimit {
				t.Fatalf("Normalized(%+v) = %+v, want page=%d limit=%d", tc.input, got, tc.wantPage, tc.wantLimit)
			}
		})
	}
}

func TestPageRequestOffset(t *testing.T) {
	if got := (PageRequest{Page: 3, Limit: 20}).Offset(); got != 40 {
		t.Fatalf("offset page=3 limit=20 = %d, want 40", got)
	}
	if got := (PageRequest{Page: 0, Limit: 20}).Offset(); got != 0 {
		t.Fatalf("offset page=0 limit=20 = %d, want 0", got)
	}
}

func TestNewPageResponse(t *testing.T) {
	resp := NewPageResponse([]string{"a", "b"}, 55, 3, 20)
	if resp.Total != 55 || resp.Page != 3 {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp.TotalPages != 3 {
		t.Fatalf("TotalPages = %d, want 3", resp.TotalPages)
	}
}
