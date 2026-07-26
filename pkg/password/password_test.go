package password

import "testing"

func TestHashAndCompare(t *testing.T) {
	hashed, err := Hash("secret123", 4)
	if err != nil {
		t.Fatalf("Hash returned error: %v", err)
	}
	if hashed == "" || hashed == "secret123" {
		t.Fatalf("Hash returned unexpected value %q", hashed)
	}
	if err := Compare(hashed, "secret123"); err != nil {
		t.Fatalf("Compare valid password returned error: %v", err)
	}
	if err := Compare(hashed, "wrong"); err == nil {
		t.Fatalf("Compare wrong password did not return error")
	}
}

func TestHashFallsBackToDefaultCost(t *testing.T) {
	// Out-of-range cost should not error; it should fall back to DefaultCost.
	if _, err := Hash("secret123", 0); err != nil {
		t.Fatalf("Hash with cost 0 returned error: %v", err)
	}
	if _, err := Hash("secret123", 99); err != nil {
		t.Fatalf("Hash with cost 99 returned error: %v", err)
	}
}
