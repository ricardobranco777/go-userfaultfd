/* SPDX-License-Identifier: BSD-2-Clause */

package userfaultfd

import (
	"os"
	"strconv"
	"strings"
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
