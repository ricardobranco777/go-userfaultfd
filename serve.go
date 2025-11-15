/* SPDX-License-Identifier: BSD-2-Clause */

package userfaultfd

import (
	"io"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Serve creates an anonymous page-fault-backed mapping,
// registers it with UFFD, and starts a goroutine to serve page faults.
// It returns the mapping []byte, a cleanup function that waits for Serve to exit
func Serve(r io.ReaderAt, size int) ([]byte, func() error, error) {
	u, err := New(UFFD_USER_MODE_ONLY, 0)
	if err != nil {
		return nil, nil, err
	}

	length := (size + pageSize - 1) &^ (pageSize - 1)

	mem, err := unix.Mmap(-1, 0, length, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_PRIVATE|unix.MAP_ANONYMOUS)
	if err != nil {
		_ = u.Close()
		return nil, nil, err
	}

	start := uintptr(unsafe.Pointer(&mem[0]))

	if _, err := u.Register(start, length, UFFDIO_REGISTER_MODE_MISSING); err != nil {
		_ = unix.Munmap(mem)
		_ = u.Close()
		return nil, nil, err
	}

	go func() {
		_ = u.Serve(start, r)
	}()

	cleanup := func() error {
		_ = u.Unregister(start, length)
		_ = unix.Munmap(mem)
		_ = u.Close()
		return nil
	}

	return mem[:size], cleanup, nil
}
