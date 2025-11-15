/* SPDX-License-Identifier: BSD-2-Clause */

package userfaultfd

import (
	"errors"
	"fmt"
	"strings"

	"golang.org/x/sys/unix"
)

var (
	ErrInvalidApi         = errors.New("kernel returned unexpected UFFD_API version")
	ErrMissingIoctl       = errors.New("missing ioctl")
	ErrUnsupportedFeature = errors.New("requested userfaultfd features not supported by kernel")
)

// PollError indicates a poll(2) error condition such as POLLERR, POLLHUP, or POLLNVAL.
type PollError struct {
	Revents int16
}

// ReventString converts the poll event mask into a human-readable flag list.
func ReventString(revents int16) string {
	var parts []string

	if revents&unix.POLLIN != 0 {
		parts = append(parts, "POLLIN")
	}
	if revents&unix.POLLOUT != 0 {
		parts = append(parts, "POLLOUT")
	}
	if revents&unix.POLLERR != 0 {
		parts = append(parts, "POLLERR")
	}
	if revents&unix.POLLHUP != 0 {
		parts = append(parts, "POLLHUP")
	}
	if revents&unix.POLLNVAL != 0 {
		parts = append(parts, "POLLNVAL")
	}

	if len(parts) == 0 {
		return fmt.Sprintf("0x%x", revents)
	}
	return strings.Join(parts, "|")
}

func (e *PollError) Error() string {
	return fmt.Sprintf("poll error: %s", ReventString(e.Revents))
}

func (e *PollError) Is(target error) bool {
	_, ok := target.(*PollError)
	return ok
}

func (e *PollError) IsHangup() bool  { return e.Revents&unix.POLLHUP != 0 }
func (e *PollError) IsError() bool   { return e.Revents&unix.POLLERR != 0 }
func (e *PollError) IsInvalid() bool { return e.Revents&unix.POLLNVAL != 0 }
