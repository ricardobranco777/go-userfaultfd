/* SPDX-License-Identifier: BSD-2-Clause */

package userfaultfd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Uffd wraps a userfaultfd file descriptor.
type Uffd struct {
	File *os.File
	api  *UffdioApi
}

// New creates a new userfaultfd and performs the two-step API handshake.
// Returns an *Uffd or an error.
func New(flags int, features uint64) (*Uffd, error) {
	// Force O_NONBLOCK as we're using poll before read to detect errors
	file, err := Open(flags | unix.O_NONBLOCK)
	if err != nil {
		return nil, err
	}

	api, err := ApiHandshake(file.Fd(), 0)
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
		if file, err = Open(flags); err != nil {
			return nil, err
		}
		if api, err = ApiHandshake(file.Fd(), features); err != nil {
			file.Close()
			return nil, err
		}
	}

	return &Uffd{
		File: file,
		api:  api,
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
	return Continue(u.File.Fd(), start, length, mode)
}

// Copy resolves a page fault by copying from src to dst.
func (u *Uffd) Copy(dst, src uintptr, length int, mode int) (int64, error) {
	return Copy(u.File.Fd(), dst, src, length, mode)
}

// Move moves pages from src to dst.
func (u *Uffd) Move(dst, src uintptr, length int, mode int) (int64, error) {
	return Move(u.File.Fd(), dst, src, length, mode)
}

// Poison poisons pages in the given range.
func (u *Uffd) Poison(start uintptr, length int, mode int) (int64, error) {
	return Poison(u.File.Fd(), start, length, mode)
}

// Register registers a memory range with the given mode.
func (u *Uffd) Register(start uintptr, length int, mode int) (*UffdioRegister, error) {
	return Register(u.File.Fd(), start, length, mode)
}

// Unregister unregisters a previously registered range.
func (u *Uffd) Unregister(start uintptr, length int) error {
	return Unregister(u.File.Fd(), start, length)
}

// Wake wakes blocked page faults in the given range.
func (u *Uffd) Wake(start uintptr, length int) error {
	return Wake(u.File.Fd(), start, length)
}

// WriteProtect enables/disables write protection.
func (u *Uffd) WriteProtect(start uintptr, length int, mode int) error {
	return WriteProtect(u.File.Fd(), start, length, mode)
}

// Zeropage zero-fills a memory range.
func (u *Uffd) Zeropage(start uintptr, length int, mode int) (int64, error) {
	return Zeropage(u.File.Fd(), start, length, mode)
}

// ReadMsgTimeout reads one event message from the userfaultfd.
// On POLLERR, POLLHUP, or POLLNVAL, a *PollError is returned.
func (u *Uffd) ReadMsgTimeout(timeout int) (*UffdMsg, error) {
	pfd := []unix.PollFd{{
		Fd:     int32(u.Fd()),
		Events: unix.POLLIN,
	}}

	if err := retryOnEINTR(func() error {
		_, err := unix.Poll(pfd, timeout)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, os.NewSyscallError("poll", err)
	}
	// From userfaultfd(2):
	// If the O_NONBLOCK flag is not enabled, then poll(2) (always) indicates the file as having a POLLERR condition.
	re := pfd[0].Revents
	if re&(unix.POLLERR|unix.POLLHUP|unix.POLLNVAL) != 0 {
		return nil, &PollError{Revents: re}
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

// ReadMsg reads a single event message from the userfaultfd
func (u *Uffd) ReadMsg() (*UffdMsg, error) {
	for {
		m, err := u.ReadMsgTimeout(-1)
		if errors.Is(err, unix.EAGAIN) {
			continue
		}
		return m, err
	}
}

// Serve runs a pagefault loop and resolves faults using the provided src.
// Blocks until closed or provider returns a non-nil error.
func (u *Uffd) Serve(start uintptr, src io.ReaderAt) error {
	// Use mmap to get page-aligned memory
	mem, err := unix.Mmap(-1, 0, pageSize, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_PRIVATE|unix.MAP_ANONYMOUS)
	if err != nil {
		return os.NewSyscallError("mmap", err)
	}
	defer unix.Munmap(mem)

	for {
		msg, err := u.ReadMsg()
		if err != nil {
			return err
		}
		if msg.Event != UFFD_EVENT_PAGEFAULT {
			continue
		}

		addr := uintptr(msg.GetPagefault().Address)
		// Byte offset into src
		offset := int64(addr - start)

		// Read from src into memory
		if _, err := src.ReadAt(mem, offset); err != nil && !errors.Is(err, io.EOF) {
			return fmt.Errorf("reader error at offset %d: %w", offset, err)
		}

		// Copy or Move to destination
		// TODO Add benchmark for Move vs Copy
		if HaveIoctlMove {
			_, err = u.Move(addr, uintptr(unsafe.Pointer(&mem[0])), pageSize, 0)
		} else {
			_, err = u.Copy(addr, uintptr(unsafe.Pointer(&mem[0])), pageSize, 0)
		}
		if err != nil {
			// Page already mapped
			if errors.Is(err, unix.EEXIST) {
				continue
			}
			return fmt.Errorf("Move failed at 0x%x offset %d: %w", addr, offset, err)
		}
	}
}
