package tui

type textStyle struct {
	start string
	end   string
}

func (s textStyle) Render(text string) string {
	if s.start == "" {
		return text
	}
	return s.start + text + s.end
}

const ansiReset = "\x1b[0m"

var (
	selectedStyle = textStyle{start: "\x1b[1m", end: ansiReset}         // bold
	activeStyle   = textStyle{start: "\x1b[4m", end: ansiReset}         // underline
	hintStyle     = textStyle{start: "\x1b[2m", end: ansiReset}         // faint
	missingStyle  = textStyle{start: "\x1b[31m\x1b[1m", end: ansiReset} // red + bold
	dirtyStyle    = textStyle{start: "\x1b[1m", end: ansiReset}         // bold
)
