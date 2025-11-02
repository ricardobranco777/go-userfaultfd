/* SPDX-License-Identifier: BSD-2-Clause */

// Package userfaultfd provides a thin wrapper around Linux's userfaultfd(2) API.
// It allows userland page-fault handling via ioctls defined in <linux/userfaultfd.h>.
package userfaultfd

import (
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

func ioctl(fd int, op uintptr, arg unsafe.Pointer) error {
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), uintptr(op), uintptr(arg))
	if errno != 0 {
		return os.NewSyscallError("ioctl", errno)
	}
	return nil
}

// NewFile creates a new userfaultfd file descriptor with the given flags.
// Returns an *os.File wrapping the fd or an error.
func NewFile(flags int) (*os.File, error) {
	fd, _, errno := unix.Syscall(uintptr(unix.SYS_USERFAULTFD), uintptr(flags), 0, 0)
	if errno != 0 {
		return nil, os.NewSyscallError("userfaultfd", errno)
	}
	return os.NewFile(fd, "userfaultfd"), nil
}

// NewFile2 creates a a new userfaultfd file descriptor with the given flags.
// Returns an *os.File wrapping the fd or an error.
func NewFile2(flags int) (*os.File, error) {
	dev, err := os.OpenFile("/dev/userfaultfd", os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	defer dev.Close()

	// ioctl(fd, USERFAULTFD_IOC_NEW, flags)
	fd, _, errno := unix.Syscall(unix.SYS_IOCTL, dev.Fd(), uintptr(USERFAULTFD_IOC_NEW), uintptr(flags))
	if errno != 0 {
		return nil, os.NewSyscallError("ioctl(USERFAULTFD_IOC_NEW)", errno)
	}
	return os.NewFile(fd, "userfaultfd"), nil
}

// ApiHandshake negotiates the userfaultfd API version and features.
// Returns the negotiated info or an error.
func ApiHandshake(fd int, features uint64) (UffdioApi, error) {
	api := UffdioApi{Api: UFFD_API, Features: features}
	if err := ioctl(fd, UFFDIO_API, unsafe.Pointer(&api)); err != nil {
		return UffdioApi{}, err
	}
	return api, nil
}

// Continue resolves a minor page fault for the given range.
func Continue(fd int, start, length uintptr, mode uint64) error {
	if !HaveIoctlContinue {
		return ErrMissingIoctl
	}
	c := UffdioContinue{
		Range: UffdioRange{
			Start: uint64(start),
			Len:   uint64(length),
		},
		Mode: mode,
	}
	if err := ioctl(fd, UFFDIO_CONTINUE, unsafe.Pointer(&c)); err != nil {
		return err
	}
	return nil
}

// Copy resolves a page fault by copying content from src to dst.
// Returns the number of bytes copied or an error.
func Copy(fd int, dst, src, length uintptr, mode uint64) (int64, error) {
	c := UffdioCopy{Dst: uint64(dst), Src: uint64(src), Len: uint64(length), Mode: mode}
	if err := ioctl(fd, UFFDIO_COPY, unsafe.Pointer(&c)); err != nil {
		return 0, err
	}
	return c.Copy, nil
}

// Move moves pages from src to dst within the same process.
// Returns the number of bytes/pages moved or an error.
func Move(fd int, dst, src, length uintptr, mode uint64) (int64, error) {
	if !HaveIoctlMove {
		return 0, ErrMissingIoctl
	}
	m := UffdioMove{Dst: uint64(dst), Src: uint64(src), Len: uint64(length), Mode: mode}
	if err := ioctl(fd, UFFDIO_MOVE, unsafe.Pointer(&m)); err != nil {
		return 0, err
	}
	return m.Move, nil
}

// Poison marks pages in the given range as poisoned. Subsequent accesses
// may generate SIGBUS or other behaviour depending on kernel semantics.
// Returns the number of pages/bytes updated (as reported by the kernel).
func Poison(fd int, start, length uintptr, mode uint64) (int64, error) {
	if !HaveIoctlPoison {
		return 0, ErrMissingIoctl
	}
	p := UffdioPoison{
		Range: UffdioRange{
			Start: uint64(start),
			Len:   uint64(length),
		},
		Mode: mode,
	}
	if err := ioctl(fd, UFFDIO_POISON, unsafe.Pointer(&p)); err != nil {
		return 0, err
	}
	return p.Updated, nil
}

// Register registers a memory range for userfaultfd handling with the specified mode.
// Returns the registration info or an error.
func Register(fd int, start, length uintptr, mode uint64) (*UffdioRegister, error) {
	reg := UffdioRegister{
		Range: UffdioRange{Start: uint64(start), Len: uint64(length)},
		Mode:  mode,
	}
	if err := ioctl(fd, UFFDIO_REGISTER, unsafe.Pointer(&reg)); err != nil {
		return nil, err
	}
	return &reg, nil
}

// Unregister unregisters a previously registered range.
func Unregister(fd int, start, length uintptr) error {
	r := UffdioRange{Start: uint64(start), Len: uint64(length)}
	if err := ioctl(fd, UFFDIO_UNREGISTER, unsafe.Pointer(&r)); err != nil {
		return err
	}
	return nil
}

// Wake wakes up blocked page faults in the given range.
func Wake(fd int, start, length uintptr) error {
	r := UffdioRange{Start: uint64(start), Len: uint64(length)}
	if err := ioctl(fd, UFFDIO_WAKE, unsafe.Pointer(&r)); err != nil {
		return err
	}
	return nil
}

// WriteProtect enables or disables write protection on a range.
func WriteProtect(fd int, start, length uintptr, mode uint64) error {
	if !HaveIoctlWriteProtect {
		return ErrMissingIoctl
	}
	wp := UffdioWriteprotect{Range: UffdioRange{Start: uint64(start), Len: uint64(length)}, Mode: mode}
	if err := ioctl(fd, UFFDIO_WRITEPROTECT, unsafe.Pointer(&wp)); err != nil {
		return err
	}
	return nil
}

// Zeropage resolves a page fault by zero-filling the memory range.
// Returns the length zeroed or an error.
func Zeropage(fd int, start, length uintptr, mode uint64) (int64, error) {
	z := UffdioZeropage{Range: UffdioRange{Start: uint64(start), Len: uint64(length)}, Mode: mode}
	if err := ioctl(fd, UFFDIO_ZEROPAGE, unsafe.Pointer(&z)); err != nil {
		return 0, err
	}
	return z.Zeropage, nil
}
