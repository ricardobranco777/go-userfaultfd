/* SPDX-License-Identifier: BSD-2-Clause */

package userfaultfd

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Uffd wraps a userfaultfd file descriptor.
type Uffd struct {
	File  *os.File
	api   *UffdioApi
	flags int
}

// Force non-blocking so we can use poll()
// Also force close-on-exec
const force = unix.O_NONBLOCK | unix.O_CLOEXEC

// New creates a new userfaultfd and performs the two-step API handshake.
// Returns an *Uffd or an error.
func New(flags int, features uint64) (*Uffd, error) {
	flags |= force
	file, err := NewFile(flags)
	if err != nil {
		return nil, err
	}
	return newCommon(file, flags, features)
}

// NewDevUserfaultfd creates a new userfaultfd and performs the two-step API handshake.
// Returns an *Uffd or an error.
func New2(flags int, features uint64) (*Uffd, error) {
	flags |= force
	file, err := NewFile2(flags)
	if err != nil {
		return nil, err
	}
	return newCommon(file, flags, features)
}

func newCommon(file *os.File, flags int, features uint64) (*Uffd, error) {
	api, err := ApiHandshake(int(file.Fd()), 0)
	if err != nil {
		file.Close()
		return nil, err
	}

	if api.Api != UFFD_API {
		return nil, ErrInvalidApi
	}

	// From UFFDIO_API(2) BUGS section:
	// In order to detect available userfault features and enable some subset of those features
	// the userfaultfd file descriptor must be closed after the first UFFDIO_API operation that
	// queries features availability and reopened before the second UFFDIO_API operation that
	// actually enables the desired features.
	if features != 0 {
		file.Close()
		if api.Features&features != features {
			return nil, ErrUnsupportedFeature
		}
		if file, err = NewFile(flags); err != nil {
			return nil, err
		}
		if api, err = ApiHandshake(int(file.Fd()), features); err != nil {
			file.Close()
			return nil, err
		}
	}

	return &Uffd{
		File:  file,
		api:   api,
		flags: flags,
	}, nil
}

// Close closes the underlying file descriptor.
func (u *Uffd) Close() error {
	return u.File.Close()
}

// FD returns the underlying file descriptor.
func (u *Uffd) Fd() int {
	return int(u.File.Fd())
}

// Features returns the API features.
func (u *Uffd) Features() uint64 {
	return u.api.Features
}

// Return the ioctls.
func (u *Uffd) Ioctls() uint64 {
	return u.api.Ioctls
}

// Returns string representation.
func (u *Uffd) String() string {
	return fmt.Sprintf("uffd(fd=%d, features=%#x, ioctls=%#x)", u.Fd(), u.api.Features, u.api.Ioctls)
}

// Returns true if ioctl is available.
func (u *Uffd) HasIoctl(ioctl int) bool {
	return ioctl != -1 && u.api.Ioctls&(1<<ioctl) != 0
}

// Continue resolves a minor page fault.
func (u *Uffd) Continue(start uintptr, length int, mode int) error {
	return Continue(u.Fd(), start, length, mode)
}

// Copy resolves a page fault by copying from src to dst.
func (u *Uffd) Copy(dst, src uintptr, length int, mode int) (int64, error) {
	return Copy(u.Fd(), dst, src, length, mode)
}

// Move moves pages from src to dst.
func (u *Uffd) Move(dst, src uintptr, length int, mode int) (int64, error) {
	return Move(u.Fd(), dst, src, length, mode)
}

// Poison poisons pages in the given range.
func (u *Uffd) Poison(start uintptr, length int, mode int) (int64, error) {
	return Poison(u.Fd(), start, length, mode)
}

// Register registers a memory range with the given mode.
func (u *Uffd) Register(start uintptr, length int, mode int) (*UffdioRegister, error) {
	return Register(u.Fd(), start, length, mode)
}

// Unregister unregisters a previously registered range.
func (u *Uffd) Unregister(start uintptr, length int) error {
	return Unregister(u.Fd(), start, length)
}

// Wake wakes blocked page faults in the given range.
func (u *Uffd) Wake(start uintptr, length int) error {
	return Wake(u.Fd(), start, length)
}

// WriteProtect enables/disables write protection.
func (u *Uffd) WriteProtect(start uintptr, length int, mode int) error {
	return WriteProtect(u.Fd(), start, length, mode)
}

// Zeropage zero-fills a memory range.
func (u *Uffd) Zeropage(start uintptr, length int, mode int) (int64, error) {
	return Zeropage(u.Fd(), start, length, mode)
}

// ReadMsg reads one event message from the userfaultfd.
// If no event is available, it returns unix.EAGAIN.
// Returns the message and any read or decoding error.
func (u *Uffd) ReadMsg() (*UffdMsg, error) {
	pfd := []unix.PollFd{{
		Fd:     int32(u.Fd()),
		Events: unix.POLLIN,
	}}

	if err := retryOnEINTR(func() error {
		n, err := unix.Poll(pfd, 0)
		if err != nil {
			return err
		}
		if n == 0 {
			return unix.EAGAIN
		}

		re := pfd[0].Revents
		if re&(unix.POLLERR|unix.POLLHUP|unix.POLLNVAL) != 0 {
			return fmt.Errorf("poll error: revents=%#x", re)
		}
		return nil
	}); err != nil {
		return nil, os.NewSyscallError("poll", err)
	}

	var msg UffdMsg
	buf := (*[unsafe.Sizeof(msg)]byte)(unsafe.Pointer(&msg))[:]

	if err := retryOnEINTR(func() error {
		n, err := unix.Read(u.Fd(), buf)
		if err != nil {
			return err
		}
		if n != len(buf) {
			return fmt.Errorf("truncated read: got %d, expected %d", n, len(buf))
		}
		return nil
	}); err != nil {
		return nil, os.NewSyscallError("read", err)
	}

	return &msg, nil
}
