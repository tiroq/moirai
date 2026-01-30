package tea

import (
	"bufio"
	"fmt"
	"io"
	"unicode/utf8"
)

// KeyType describes the kind of key pressed.
type KeyType int

const (
	KeyUnknown KeyType = iota
	KeyRunes
	KeyEnter
	KeyEsc
	KeyBackspace
	KeyDelete
	KeyCtrlC
	KeyUp
	KeyDown
	KeyPgUp
	KeyPgDown
)

// KeyMsg represents a keyboard event.
type KeyMsg struct {
	Type  KeyType
	Runes []rune
}

func (k KeyMsg) String() string {
	switch k.Type {
	case KeyRunes:
		return string(k.Runes)
	case KeyEnter:
		return "enter"
	case KeyEsc:
		return "esc"
	case KeyBackspace:
		return "backspace"
	case KeyDelete:
		return "delete"
	case KeyCtrlC:
		return "ctrl+c"
	case KeyUp:
		return "up"
	case KeyDown:
		return "down"
	case KeyPgUp:
		return "pgup"
	case KeyPgDown:
		return "pgdown"
	default:
		return ""
	}
}

// WindowSizeMsg describes a terminal window size.
type WindowSizeMsg struct {
	Width  int
	Height int
}

func readKeyMsg(r *bufio.Reader) (Msg, error) {
	b, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	switch b {
	case 0x03:
		return KeyMsg{Type: KeyCtrlC}, nil
	case '\r', '\n':
		return KeyMsg{Type: KeyEnter}, nil
	case 0x7f:
		return KeyMsg{Type: KeyBackspace}, nil
	case 0x1b:
		return readEscapeSequence(r)
	default:
		if b < utf8.RuneSelf {
			return KeyMsg{Type: KeyRunes, Runes: []rune{rune(b)}}, nil
		}
		// Multi-byte UTF-8 rune; read the remainder.
		buf := []byte{b}
		for !utf8.FullRune(buf) {
			next, err := r.ReadByte()
			if err != nil {
				return nil, err
			}
			buf = append(buf, next)
			if len(buf) > 4 {
				break
			}
		}
		rn, _ := utf8.DecodeRune(buf)
		if rn == utf8.RuneError {
			return KeyMsg{Type: KeyUnknown}, nil
		}
		return KeyMsg{Type: KeyRunes, Runes: []rune{rn}}, nil
	}
}

func readEscapeSequence(r *bufio.Reader) (Msg, error) {
	peek, err := r.Peek(1)
	if err != nil {
		if err == io.EOF {
			return KeyMsg{Type: KeyEsc}, nil
		}
		return nil, err
	}
	if peek[0] != '[' {
		return KeyMsg{Type: KeyEsc}, nil
	}
	_, _ = r.ReadByte() // '['

	code, err := r.ReadByte()
	if err != nil {
		return nil, err
	}

	switch code {
	case 'A':
		return KeyMsg{Type: KeyUp}, nil
	case 'B':
		return KeyMsg{Type: KeyDown}, nil
	case '5':
		if tilde, err := r.ReadByte(); err == nil && tilde == '~' {
			return KeyMsg{Type: KeyPgUp}, nil
		}
		return KeyMsg{Type: KeyUnknown}, nil
	case '6':
		if tilde, err := r.ReadByte(); err == nil && tilde == '~' {
			return KeyMsg{Type: KeyPgDown}, nil
		}
		return KeyMsg{Type: KeyUnknown}, nil
	default:
		return KeyMsg{Type: KeyEsc}, nil
	}
}

func (k KeyType) GoString() string { return fmt.Sprintf("KeyType(%d)", int(k)) }
