// +build darwin linux freebsd openbsd netbsd

package log

import (
	"os"
	"syscall"
)

func Dup2File(file *os.File, fd int) error {
	if err := syscall.Dup2(int(file.Fd()), fd); err != nil {
		return err
	}
	return nil
}
