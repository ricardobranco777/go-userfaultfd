/* SPDX-License-Identifier: BSD-2-Clause */

package userfaultfd

import (
	"errors"
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
		t.Fatalf("expected EAGAIN or EWOULDBLOCK, got nil")
	}
	if !errors.Is(err, unix.EAGAIN) && !errors.Is(err, unix.EWOULDBLOCK) {
		t.Fatalf("unexpected error: %v", err)
	}
}
