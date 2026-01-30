//go:build !linux

package tea

import "os"

func windowSize(_ *os.File) (width, height int, ok bool) {
	return 0, 0, false
}

