/* SPDX-License-Identifier: BSD-2-Clause */

package userfaultfd

import (
	"errors"
	"testing"

	"golang.org/x/sys/unix"
)

func TestPollErrorImplementsError(t *testing.T) {
	err := &PollError{Revents: unix.POLLERR}
	var e error = err
	if e.Error() == "" {
		t.Fatalf("expected non-empty error string")
	}
}

func TestPollErrorIs(t *testing.T) {
	err := &PollError{Revents: unix.POLLERR}

	// Must satisfy errors.Is for same type regardless of instance
	if !errors.Is(err, &PollError{}) {
		t.Fatalf("expected errors.Is to match PollError type")
	}

	// Must not match unrelated errors
	if errors.Is(err, ErrMissingIoctl) {
		t.Fatalf("unexpected match against unrelated error")
	}
}

func TestPollErrorHelpers(t *testing.T) {
	e := &PollError{Revents: unix.POLLERR | unix.POLLHUP}

	if !e.IsError() {
		t.Fatalf("IsError expected true")
	}
	if !e.IsHangup() {
		t.Fatalf("IsHangup expected true")
	}
	if e.IsInvalid() {
		t.Fatalf("IsInvalid expected false")
	}
}

func TestReventStringSingle(t *testing.T) {
	cases := []struct {
		rev  int16
		want string
	}{
		{unix.POLLIN, "POLLIN"},
		{unix.POLLOUT, "POLLOUT"},
		{unix.POLLERR, "POLLERR"},
		{unix.POLLHUP, "POLLHUP"},
		{unix.POLLNVAL, "POLLNVAL"},
	}

	for _, tc := range cases {
		got := ReventString(tc.rev)
		if got != tc.want {
			t.Fatalf("ReventString(%#x) = %q, want %q", tc.rev, got, tc.want)
		}
	}
}

func TestReventStringMultiple(t *testing.T) {
	flags := int16(unix.POLLERR | unix.POLLHUP | unix.POLLOUT)
	got := ReventString(flags)
	want := "POLLOUT|POLLERR|POLLHUP" // ordering same as implementation
	if got != want {
		t.Fatalf("ReventString combined flags = %q, want %q", got, want)
	}
}

func TestReventStringZero(t *testing.T) {
	got := ReventString(0)
	want := "0x0"
	if got != want {
		t.Fatalf("ReventString(0) = %q, want %q", got, want)
	}
}

func TestPollErrorErrorFormat(t *testing.T) {
	e := &PollError{Revents: unix.POLLERR | unix.POLLNVAL}
	got := e.Error()
	want := "poll error: POLLERR|POLLNVAL"
	if got != want {
		t.Fatalf("Error() = %q, want %q", got, want)
	}
}
