/* SPDX-License-Identifier: BSD-2-Clause */

package userfaultfd

import (
	"errors"
	"os"
	"testing"

	"golang.org/x/sys/unix"
)

func TestNew(t *testing.T) {
	// Try creating a userfaultfd with no special features
	uffd, err := New(flags, 0)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}

	if uffd.Fd() < 0 {
		t.Errorf("invalid fd: %d", uffd.Fd())
	}

	fdFlags, _ := unix.FcntlInt(uintptr(uffd.Fd()), unix.F_GETFD, 0)
	if fdFlags&unix.FD_CLOEXEC == 0 {
		t.Errorf("FD_CLOEXEC not set")
	}

	fl, _ := unix.FcntlInt(uintptr(uffd.Fd()), unix.F_GETFL, 0)
	if fl&unix.O_NONBLOCK == 0 {
		t.Errorf("O_NONBLOCK not set")
	}

	uffd.Close()
	if err := unix.Close(uffd.Fd()); err == nil {
		t.Fatal("Close failed")
	}

	t.Logf("Userfaultfd API: %d, features: 0x%x, ioctls: 0x%x", uffd.api.Api, uffd.Features(), uffd.Ioctls())

	// Test enabling a feature
	features := uint64(UFFD_FEATURE_PAGEFAULT_FLAG_WP)
	uffd, err = New(flags, features)
	if err != nil {
		t.Logf("New with requested features 0x%x skipped: %v", features, err)
	} else {
		uffd.Close()
	}
}

func TestNew2(t *testing.T) {
	if !HaveDevUserfaultfd {
		t.Skip("/dev/userfaultfd does not exist")
	}
	// Try creating a userfaultfd with no special features
	uffd, err := New2(flags, 0)
	if err != nil {
		if errors.Is(err, unix.EACCES) {
			t.Skip("/dev/userfaultfd is not readable")
		} else {
			t.Fatalf("NewFile2 failed: %v", err)
		}
	}

	if uffd.Fd() < 0 {
		t.Errorf("invalid fd: %d", uffd.Fd())
	}

	uffd.Close()
	if err := unix.Close(uffd.Fd()); err == nil {
		t.Fatal("Close failed")
	}

	t.Logf("Userfaultfd API: %d, features: 0x%x, ioctls: 0x%x", uffd.api.Api, uffd.Features(), uffd.Ioctls())

	// Test enabling a feature
	features := uint64(UFFD_FEATURE_PAGEFAULT_FLAG_WP)
	uffd, err = New2(flags, features)
	if err != nil {
		t.Logf("New with requested features 0x%x skipped: %v", features, err)
	} else {
		uffd.Close()
	}
}

func TestReadMsgNoEvent(t *testing.T) {
	uffd, err := New(flags|unix.O_NONBLOCK, 0)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer uffd.Close()

	_, err = uffd.ReadMsg()
	if err == nil {
		t.Fatalf("expected EAGAIN, got nil")
	}
	if !errors.Is(err, unix.EAGAIN) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReadMsgNonBlocking(t *testing.T) {
	uffd, err := New(flags|unix.O_NONBLOCK, 0)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer uffd.Close()

	// Explicitly verify polling behavior inside ReadMsg() with non-blocking FD
	_, err = uffd.ReadMsg()
	if err == nil {
		t.Fatalf("expected EAGAIN from nonblocking read, got nil")
	}

	// ReadMsg wraps read errors in os.NewSyscallError("read", err)
	var serr *os.SyscallError
	if !errors.As(err, &serr) {
		t.Fatalf("expected *os.SyscallError wrapping EAGAIN, got %T: %v", err, err)
	}

	if !errors.Is(serr.Err, unix.EAGAIN) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHasIoctl(t *testing.T) {
	uffd, err := New(flags|unix.O_NONBLOCK, 0)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer uffd.Close()

	tests := []struct {
		ioctl int
		want  bool
	}{
		{_UFFDIO_API, true},
		{_UFFDIO_REGISTER, true},
		{_UFFDIO_UNREGISTER, true},
	}

	for _, tt := range tests {
		got := uffd.HasIoctl(tt.ioctl)
		if got != tt.want {
			t.Fatalf("HasIoctl(%d) = %v, want %v", tt.ioctl, got, tt.want)
		}
	}
}
