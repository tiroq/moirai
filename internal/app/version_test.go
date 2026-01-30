package app

import "testing"

func TestDefaultVersionIsDev(t *testing.T) {
	if Version != "dev" {
		t.Fatalf("expected default Version to be %q, got %q", "dev", Version)
	}
}
