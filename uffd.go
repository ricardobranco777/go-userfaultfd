/* SPDX-License-Identifier: BSD-2-Clause */

package userfaultfd

import (
	"os"
)

// Uffd wraps a userfaultfd file descriptor.
type Uffd struct {
	File  *os.File
	api   UffdioApi
	flags int
}

// New creates a new userfaultfd and performs the two-step API handshake.
func New(flags int, features uint64) (*Uffd, error) {
	file, err := NewFile(flags)
	if err != nil {
		return nil, err
	}

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

// Continue resolves a minor page fault.
func (u *Uffd) Continue(start, length uintptr, mode uint64) error {
	return Continue(u.Fd(), start, length, mode)
}

// Copy resolves a page fault by copying from src to dst.
func (u *Uffd) Copy(dst, src, length uintptr, mode uint64) (int64, error) {
	return Copy(u.Fd(), dst, src, length, mode)
}

// Move moves pages from src to dst.
func (u *Uffd) Move(dst, src, length uintptr, mode uint64) (int64, error) {
	return Move(u.Fd(), dst, src, length, mode)
}

// Poison poisons pages in the given range.
func (u *Uffd) Poison(start, length uintptr, mode uint64) (int64, error) {
	return Poison(u.Fd(), start, length, mode)
}

// Register registers a memory range with the given mode.
func (u *Uffd) Register(start, length uintptr, mode uint64) (*UffdioRegister, error) {
	return Register(u.Fd(), start, length, mode)
}

// Unregister unregisters a previously registered range.
func (u *Uffd) Unregister(start, length uintptr) error {
	return Unregister(u.Fd(), start, length)
}

// Wake wakes blocked page faults in the given range.
func (u *Uffd) Wake(start, length uintptr) error {
	return Wake(u.Fd(), start, length)
}

// WriteProtect enables/disables write protection.
func (u *Uffd) WriteProtect(start, length uintptr, mode uint64) error {
	return WriteProtect(u.Fd(), start, length, mode)
}

// Zeropage zero-fills a memory range.
func (u *Uffd) Zeropage(start, length uintptr, mode uint64) (int64, error) {
	return Zeropage(u.Fd(), start, length, mode)
}
