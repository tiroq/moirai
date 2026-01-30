//go:build !windows && !linux

package tea

import "os"

func enterRawMode(_ *os.File) (func(), error) {
	return func() {}, nil
}
