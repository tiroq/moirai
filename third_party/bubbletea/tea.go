package tea

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"
)

// Msg represents a message passed to a model update.
type Msg any

// Cmd performs work and returns a Msg.
type Cmd func() Msg

// Model is the core Bubble Tea interface.
type Model interface {
	Init() Cmd
	Update(Msg) (Model, Cmd)
	View() string
}

// QuitMsg signals the program should exit.
type QuitMsg struct{}

// Quit is a command that terminates a program.
func Quit() Msg { return QuitMsg{} }

// Program runs a model.
type Program struct {
	model Model
}

// NewProgram creates a new program for the given model.
func NewProgram(m Model) *Program {
	return &Program{model: m}
}

// Run runs the program until it quits.
func (p *Program) Run() (Model, error) {
	if p.model == nil {
		return nil, fmt.Errorf("model is required")
	}

	m := p.model
	reader := bufio.NewReader(os.Stdin)

	restore, err := enterRawMode(os.Stdin)
	if err != nil {
		return nil, err
	}
	defer restore()

	if w, h, ok := windowSize(os.Stdout); ok {
		if next, cmd := m.Update(WindowSizeMsg{Width: w, Height: h}); next != nil {
			m = next
			// Ignore cmd; size updates shouldn't trigger side effects.
			_ = cmd
		}
	}

	clearScreen(os.Stdout)
	if _, err := io.WriteString(os.Stdout, m.View()); err != nil {
		return nil, err
	}

	cmd := m.Init()
	if cmd != nil {
		var quit bool
		m, quit, err = runCmdChain(m, cmd)
		if err != nil {
			return nil, err
		}
		if quit {
			return m, nil
		}
		clearScreen(os.Stdout)
		if _, err := io.WriteString(os.Stdout, m.View()); err != nil {
			return nil, err
		}
	}

	for {
		msg, err := readKeyMsg(reader)
		if err != nil {
			return m, err
		}

		var nextCmd Cmd
		m, nextCmd = m.Update(msg)

		if nextCmd != nil {
			var quit bool
			m, quit, err = runCmdChain(m, nextCmd)
			if err != nil {
				return m, err
			}
			if quit {
				return m, nil
			}
		}

		if w, h, ok := windowSize(os.Stdout); ok {
			if next, cmd := m.Update(WindowSizeMsg{Width: w, Height: h}); next != nil {
				m = next
				_ = cmd
			}
		}

		clearScreen(os.Stdout)
		if _, err := io.WriteString(os.Stdout, m.View()); err != nil {
			return m, err
		}

		// Prevent busy loops if the terminal doesn't deliver key events.
		if runtime.GOOS == "windows" {
			_, _ = os.Stdout.WriteString("")
		}
	}
}

func runCmdChain(m Model, cmd Cmd) (Model, bool, error) {
	current := cmd
	for current != nil {
		msg := current()
		if _, ok := msg.(QuitMsg); ok {
			return m, true, nil
		}
		var next Cmd
		m, next = m.Update(msg)
		current = next
	}
	return m, false, nil
}

func clearScreen(w io.Writer) {
	_, _ = fmt.Fprint(w, "\x1b[2J\x1b[H")
}
