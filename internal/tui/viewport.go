package tui

import "strings"

type diffViewport struct {
	content string
	lines   []string
	width   int
	height  int
	y       int
}

func newDiffViewport() diffViewport {
	return diffViewport{
		width:  1,
		height: 1,
	}
}

func (v *diffViewport) SetContent(content string) {
	v.content = content
	v.lines = strings.Split(content, "\n")
	v.clamp()
}

func (v *diffViewport) View() string {
	if v.height <= 0 || len(v.lines) == 0 {
		return ""
	}
	start := v.y
	end := v.y + v.height
	if end > len(v.lines) {
		end = len(v.lines)
	}
	return strings.Join(v.lines[start:end], "\n")
}

func (v *diffViewport) LineDown(lines int) {
	v.y += lines
	v.clamp()
}

func (v *diffViewport) LineUp(lines int) {
	v.y -= lines
	v.clamp()
}

func (v *diffViewport) PageDown() {
	v.y += v.height
	v.clamp()
}

func (v *diffViewport) PageUp() {
	v.y -= v.height
	v.clamp()
}

func (v *diffViewport) GotoTop() {
	v.y = 0
}

func (v *diffViewport) clamp() {
	if v.height <= 0 {
		v.y = 0
		return
	}
	max := len(v.lines) - v.height
	if max < 0 {
		max = 0
	}
	if v.y < 0 {
		v.y = 0
	}
	if v.y > max {
		v.y = max
	}
}
