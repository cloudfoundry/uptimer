package syscallshim

import "syscall"

func (sh *SyscallShim) Faccessat(dirfd int, path string, mode uint32, flags int) (err error) {
	return syscall.Faccessat(dirfd, path, mode, flags)
}

