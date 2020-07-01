package syscallshim

func (sh *SyscallShim) Faccessat(dirfd int, path string, mode uint32, flags int) (err error) {
	panic("unsupported!")
}

