package tea

import (
	"bufio"
	"bytes"
	"testing"
)

func TestReadKeyMsgEscDoesNotBlock(t *testing.T) {
	r := bufio.NewReader(bytes.NewReader([]byte{0x1b}))
	msg, err := readKeyMsg(r)
	if err != nil {
		t.Fatalf("readKeyMsg: %v", err)
	}
	km, ok := msg.(KeyMsg)
	if !ok {
		t.Fatalf("expected KeyMsg, got %T", msg)
	}
	if km.Type != KeyEsc {
		t.Fatalf("expected KeyEsc, got %v", km.Type)
	}
}

func TestReadKeyMsgArrowUp(t *testing.T) {
	r := bufio.NewReader(bytes.NewReader([]byte{0x1b, '[', 'A'}))
	msg, err := readKeyMsg(r)
	if err != nil {
		t.Fatalf("readKeyMsg: %v", err)
	}
	km, ok := msg.(KeyMsg)
	if !ok {
		t.Fatalf("expected KeyMsg, got %T", msg)
	}
	if km.Type != KeyUp {
		t.Fatalf("expected KeyUp, got %v", km.Type)
	}
}

