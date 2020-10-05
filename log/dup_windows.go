// +build windows

package log

import (
	"os"
	"syscall"
)

var (
	kernel32         = syscall.MustLoadDLL("kernel32.dll")
	procSetStdHandle = kernel32.MustFindProc("SetStdHandle")
)

func Dup2File(file *os.File, fd int) error {
	stdHandle := syscall.STD_ERROR_HANDLE
	if fd == 1 {
		stdHandle = syscall.STD_OUTPUT_HANDLE
	}
	err := setStdHandle(stdHandle, syscall.Handle(file.Fd()))
	if err != nil {
		return err
	}
	if fd == 1 {
		os.Stdout = file
	} else {
		os.Stderr = file
	}
	return nil
}

func setStdHandle(stdhandle int32, handle syscall.Handle) error {
	r0, _, e1 := syscall.Syscall(procSetStdHandle.Addr(), 2, uintptr(stdhandle), uintptr(handle), 0)
	if r0 == 0 {
		if e1 != 0 {
			return error(e1)
		}
		return syscall.EINVAL
	}
	return nil
}
