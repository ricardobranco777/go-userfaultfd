/* SPDX-License-Identifier: BSD-2-Clause */

package userfaultfd

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"io"
	"os"
	"testing"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

func TestNew(t *testing.T) {
	// Try creating a userfaultfd with no special features
	uffd, err := New(flags|unix.O_CLOEXEC, 0)
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

func TestReadMsgNoEvent(t *testing.T) {
	uffd, err := New(flags, 0)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer uffd.Close()

	done := make(chan struct{})
	go func() {
		// ReadMsg blocks forever now, so never returns
		_, _ = uffd.ReadMsg()
		close(done)
	}()

	select {
	case <-done:
		t.Fatalf("ReadMsg returned unexpectedly")
	case <-time.After(50 * time.Millisecond):
		// expected
	}
}

func TestReadMsgNonBlocking(t *testing.T) {
	uffd, err := New(flags, 0)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer uffd.Close()

	// Non-blocking poll attempt
	_, err = uffd.ReadMsgTimeout(0)
	if err == nil {
		t.Fatalf("expected EAGAIN from nonblocking read, got nil")
	}

	var serr *os.SyscallError
	if !errors.As(err, &serr) {
		t.Fatalf("expected *os.SyscallError wrapping EAGAIN, got %T: %v", err, err)
	}

	if !errors.Is(serr.Err, unix.EAGAIN) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHasIoctl(t *testing.T) {
	uffd, err := New(flags, 0)
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

func TestReadMsgTimeoutImmediate(t *testing.T) {
	uffd, err := New(flags, 0)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer uffd.Close()

	_, err = uffd.ReadMsgTimeout(0)
	if err == nil {
		t.Fatalf("expected EAGAIN, got nil")
	}
	if !errors.Is(err, unix.EAGAIN) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReadMsgTimeoutShort(t *testing.T) {
	uffd, err := New(flags, 0)
	if err != nil {
		t.Fatalf("New failed: %v", err)
	}
	defer uffd.Close()

	// 50ms timeout for reasonable CI runtime
	_, err = uffd.ReadMsgTimeout(50)
	if err == nil {
		t.Fatalf("expected timeout EAGAIN, got nil")
	}
	if !errors.Is(err, unix.EAGAIN) {
		t.Fatalf("unexpected error from ReadMsgTimeout(50): %v", err)
	}
}

func TestReadMsgTimeoutTable(t *testing.T) {
	tests := []struct {
		name     string
		flags    int
		timeout  int
		expectFn func(t *testing.T, err error, elapsed time.Duration)
	}{
		{
			name:    "nonblocking-timeout0",
			flags:   flags | unix.O_NONBLOCK,
			timeout: 0,
			expectFn: func(t *testing.T, err error, elapsed time.Duration) {
				if err == nil {
					t.Fatalf("expected EAGAIN, got nil")
				}
				if !errors.Is(err, unix.EAGAIN) {
					t.Fatalf("expected EAGAIN, got %v", err)
				}
				if elapsed > 10*time.Millisecond {
					t.Fatalf("nonblocking timeout=0 should return immediately, took %v", elapsed)
				}
			},
		},
		{
			name:    "nonblocking-timeout-positive",
			flags:   flags | unix.O_NONBLOCK,
			timeout: 50,
			expectFn: func(t *testing.T, err error, elapsed time.Duration) {
				if err == nil {
					t.Fatalf("expected EAGAIN, got nil")
				}
				if !errors.Is(err, unix.EAGAIN) {
					t.Fatalf("expected EAGAIN, got %v", err)
				}
				if elapsed < 45*time.Millisecond {
					t.Fatalf("timeout did not wait long enough: %v", elapsed)
				}
			},
		},
		{
			name:    "nonblocking-timeout-negative",
			flags:   flags | unix.O_NONBLOCK,
			timeout: -1,
			expectFn: func(t *testing.T, err error, elapsed time.Duration) {
				// should block forever -> must not return within test window
				if err != nil {
					t.Fatalf("expected block, got err=%v", err)
				}
				if elapsed < 40*time.Millisecond {
					t.Fatalf("should block indefinitely, returned after %v", elapsed)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uffd, err := New(tt.flags, 0)
			if err != nil {
				t.Fatalf("New failed: %v", err)
			}
			defer uffd.Close()

			done := make(chan error, 1)
			start := time.Now()

			go func() {
				_, err := uffd.ReadMsgTimeout(tt.timeout)
				done <- err
			}()

			var resultErr error

			if tt.timeout < 0 {
				// block indefinitely: expect no result in 40ms
				select {
				case resultErr = <-done:
				case <-time.After(40 * time.Millisecond):
				}
			} else {
				select {
				case resultErr = <-done:
				case <-time.After(time.Duration(tt.timeout+20) * time.Millisecond):
					t.Fatalf("ReadMsgTimeout(%d) hung unexpectedly", tt.timeout)
				}
			}

			tt.expectFn(t, resultErr, time.Since(start))
		})
	}
}

func TestUffdWithLocalFile(t *testing.T) {
	f, err := os.Open("/bin/bash")
	if err != nil {
		t.Skipf("skipping: testdata file missing: %v", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	size := int(info.Size())

	u, err := New(flags, 0)
	if err != nil {
		t.Skipf("skipping: userfaultfd not available: %v", err)
	}
	defer u.Close()

	length := (size + pageSize - 1) &^ (pageSize - 1)

	mem, err := unix.Mmap(-1, 0, length, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_PRIVATE|unix.MAP_ANONYMOUS)
	if err != nil {
		t.Fatalf("mmap failed: %v", err)
	}
	defer unix.Munmap(mem)

	start := uintptr(unsafe.Pointer(&mem[0]))

	if _, err := u.Register(start, length, UFFDIO_REGISTER_MODE_MISSING); err != nil {
		t.Fatalf("skipping: UFFD register failed: %v", err)
	}

	go func() {
		_ = u.Serve(start, f)
	}()

	// Touch the mapping to trigger faults over the region
	data := mem[:size]

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("seek failed: %v", err)
	}
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		t.Fatalf("failed computing reference hash: %v", err)
	}
	expectedHash := h.Sum(nil)

	actualHash := sha256.Sum256(data[:size])

	if !bytes.Equal(expectedHash, actualHash[:]) {
		t.Fatalf("content mismatch: expected %x got %x", expectedHash, actualHash)
	}
}
