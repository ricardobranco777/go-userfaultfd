/* SPDX-License-Identifier: BSD-2-Clause */

package userfaultfd

import "errors"

var (
	ErrInvalidApi         = errors.New("kernel returned unexpected UFFD_API version")
	ErrMissingIoctl       = errors.New("missing ioctl")
	ErrUnsupportedFeature = errors.New("requested userfaultfd features not supported by kernel")
)
