/* SPDX License Identifier: BSD-2-Clause */

package userfaultfd

var (
	// True if /proc/sys/vm/unprivileged_userfaultfd == 1
	UnprivilegedUserfaultfd bool

	// Supports /dev/userfaultfd
	HaveDevUserfaultfd bool

	// Kernel supports user mode only flag
	HaveUserModeOnly bool

	// Detect newer UFFDIO ioctls
	HaveIoctlContinue     bool
	HaveIoctlMove         bool
	HaveIoctlPoison       bool
	HaveIoctlWriteProtect bool
)

func init() {
	// Check sysctl for unprivileged_userfaultfd
	UnprivilegedUserfaultfd = UnprivilegedUserfaultfdAllowed()

	// Detect missing ioctls (declared as 0 by Cgo if not defined)
	if UFFDIO_CONTINUE != 0 {
		HaveIoctlContinue = true
	}
	if UFFDIO_MOVE != 0 {
		HaveIoctlMove = true
	}
	if UFFDIO_POISON != 0 {
		HaveIoctlPoison = true
	}
	if UFFDIO_WRITEPROTECT != 0 {
		HaveIoctlWriteProtect = true
	}

	if UFFD_USER_MODE_ONLY != 0 {
		HaveUserModeOnly = true
	}

	if USERFAULTFD_IOC_NEW != 0 {
		HaveDevUserfaultfd = true
	}
}
