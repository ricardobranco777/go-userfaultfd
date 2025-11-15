/* SPDX-License-Identifier: BSD-2-Clause */

package userfaultfd

import (
	"os"
	"strconv"
	"strings"

	"golang.org/x/sys/unix"
)

// UnprivilegedUserfaultfdAllowed returns true if
// /proc/sys/vm/unprivileged_userfaultfd contains 1
func UnprivilegedUserfaultfdAllowed() bool {
	data, err := os.ReadFile("/proc/sys/vm/unprivileged_userfaultfd")
	if err != nil {
		return false
	}
	if v, err := strconv.Atoi(strings.TrimSpace(string(data))); err != nil {
		return false
	} else {
		return v == 1
	}
}

// retryOnEINTR repeatedly calls fn until it returns nil or an error other than EINTR.
func retryOnEINTR(fn func() error) error {
	for {
		err := fn()
		if err == unix.EINTR {
			continue
		}
		return err
	}
}
