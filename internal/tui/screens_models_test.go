package tui

import "testing"

func TestModelWindow(t *testing.T) {
	start, end := modelWindow(0, 0, 10)
	if start != 0 || end != 0 {
		t.Fatalf("expected empty window, got %d-%d", start, end)
	}

	start, end = modelWindow(5, 0, 10)
	if start != 0 || end != 5 {
		t.Fatalf("expected 0-5, got %d-%d", start, end)
	}

	start, end = modelWindow(25, 0, 10)
	if start != 0 || end != 10 {
		t.Fatalf("expected 0-10, got %d-%d", start, end)
	}

	start, end = modelWindow(25, 9, 10)
	if start != 0 || end != 10 {
		t.Fatalf("expected 0-10, got %d-%d", start, end)
	}

	start, end = modelWindow(25, 10, 10)
	if start != 10 || end != 20 {
		t.Fatalf("expected 10-20, got %d-%d", start, end)
	}

	start, end = modelWindow(25, 19, 10)
	if start != 10 || end != 20 {
		t.Fatalf("expected 10-20, got %d-%d", start, end)
	}

	start, end = modelWindow(25, 20, 10)
	if start != 20 || end != 25 {
		t.Fatalf("expected 20-25, got %d-%d", start, end)
	}
}

