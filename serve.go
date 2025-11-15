/* SPDX-License-Identifier: BSD-2-Clause */

package userfaultfd

import (
	"io"
	"unsafe"

	"golang.org/x/sys/unix"
)

// ServeMapping creates an anonymous page-fault-backed mapping,
// registers it with UFFD, and starts a goroutine to serve page faults.
// It returns the mapping []byte and a Close function that waits for Serve to exit.
func ServeMapping(r io.ReaderAt, size int64) ([]byte, func() error, error) {
	pageSize := unix.Getpagesize()
	mapLen := roundUp(int(size), pageSize)

	// Create userfaultfd in non-blocking mode
	u, err := New(flags|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, nil, err
	}

	full, err := unix.Mmap(-1, 0, mapLen, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_PRIVATE|unix.MAP_ANONYMOUS)
	if err != nil {
		_ = u.Close()
		return nil, nil, err
	}

	base := uintptr(unsafe.Pointer(&full[0]))

	if _, err := u.Register(base, mapLen, UFFDIO_REGISTER_MODE_MISSING); err != nil {
		_ = unix.Munmap(full)
		_ = u.Close()
		return nil, nil, err
	}

	provider := ReaderAtPageProvider(r)

	// Start handler; it will exit when the fd is closed / mapping is gone.
	go func() {
		_ = u.Serve(base, mapLen, pageSize, provider)
	}()

	// cleanup function that waits for Serve to finish
	cleanup := func() error {
		// Best-effort cleanup; ignore Serve’s lifetime.
		_ = u.Unregister(base, mapLen)
		_ = u.Close()
		return unix.Munmap(full)
	}

	return full[:size], cleanup, nil
}
