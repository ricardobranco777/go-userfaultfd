/* SPDX-License-Identifier: BSD-2-Clause */

package userfaultfd

/*
#include <linux/ioctl.h>
#include <linux/userfaultfd.h>
#include <asm/unistd.h>

#ifndef UFFD_USER_MODE_ONLY
#define UFFD_USER_MODE_ONLY	0
#endif
#ifndef UFFDIO_CONTINUE
#define UFFDIO_CONTINUE		0
#define _UFFDIO_CONTINUE	-1
#endif
#ifndef UFFDIO_MOVE
#define UFFDIO_MOVE		0
#define _UFFDIO_MOVE		-1
#endif
#ifndef UFFDIO_POISON
#define UFFDIO_POISON		0
#define _UFFDIO_POISON		-1
#endif
#ifndef UFFDIO_WRITEPROTECT
#define UFFDIO_WRITEPROTECT	0
#define _UFFDIO_WRITEPROTECT	-1
#endif
#ifndef USERFAULTFD_IOC_NEW
#define USERFAULTFD_IOC_NEW	0
#endif
*/
import "C"

const (
	// Create a userfaultfd that can handle page faults only in user mode.
	UFFD_USER_MODE_ONLY = C.UFFD_USER_MODE_ONLY
)

const (
	UFFD_API            = C.UFFD_API
	UFFDIO_API          = C.UFFDIO_API
	UFFDIO_REGISTER     = C.UFFDIO_REGISTER
	UFFDIO_UNREGISTER   = C.UFFDIO_UNREGISTER
	UFFDIO_WAKE         = C.UFFDIO_WAKE
	UFFDIO_COPY         = C.UFFDIO_COPY
	UFFDIO_ZEROPAGE     = C.UFFDIO_ZEROPAGE
	UFFDIO_MOVE         = C.UFFDIO_MOVE
	UFFDIO_WRITEPROTECT = C.UFFDIO_WRITEPROTECT
	UFFDIO_CONTINUE     = C.UFFDIO_CONTINUE
	UFFDIO_POISON       = C.UFFDIO_POISON
	USERFAULTFD_IOC_NEW = C.USERFAULTFD_IOC_NEW
	// Used to check available Ioctls
	_UFFDIO_API          = C._UFFDIO_API
	_UFFDIO_REGISTER     = C._UFFDIO_REGISTER
	_UFFDIO_UNREGISTER   = C._UFFDIO_UNREGISTER
	_UFFDIO_WAKE         = C._UFFDIO_WAKE
	_UFFDIO_COPY         = C._UFFDIO_COPY
	_UFFDIO_ZEROPAGE     = C._UFFDIO_ZEROPAGE
	_UFFDIO_MOVE         = C._UFFDIO_MOVE
	_UFFDIO_WRITEPROTECT = C._UFFDIO_WRITEPROTECT
	_UFFDIO_CONTINUE     = C._UFFDIO_CONTINUE
	_UFFDIO_POISON       = C._UFFDIO_POISON
)

// UFFDIO_API features
const (
	UFFD_FEATURE_PAGEFAULT_FLAG_WP  = 1 << iota // 1 << 0
	UFFD_FEATURE_EVENT_FORK                     // 1 << 1
	UFFD_FEATURE_EVENT_REMAP                    // 1 << 2
	UFFD_FEATURE_EVENT_REMOVE                   // 1 << 3
	UFFD_FEATURE_MISSING_HUGETLBFS              // 1 << 4
	UFFD_FEATURE_MISSING_SHMEM                  // 1 << 5
	UFFD_FEATURE_EVENT_UNMAP                    // 1 << 6
	UFFD_FEATURE_SIGBUS                         // 1 << 7
	UFFD_FEATURE_THREAD_ID                      // 1 << 8
	UFFD_FEATURE_MINOR_HUGETLBFS                // 1 << 9
	UFFD_FEATURE_MINOR_SHMEM                    // 1 << 10
	UFFD_FEATURE_EXACT_ADDRESS                  // 1 << 11
	UFFD_FEATURE_WP_HUGETLBFS_SHMEM             // 1 << 12
	UFFD_FEATURE_WP_UNPOPULATED                 // 1 << 13
	UFFD_FEATURE_POISON                         // 1 << 14
	UFFD_FEATURE_WP_ASYNC                       // 1 << 15
	UFFD_FEATURE_MOVE                           // 1 << 16
)

// userfaultfd events
const (
	UFFD_EVENT_PAGEFAULT = 0x12
	UFFD_EVENT_FORK      = 0x13
	UFFD_EVENT_REMAP     = 0x14
	UFFD_EVENT_REMOVE    = 0x15
	UFFD_EVENT_UNMAP     = 0x16
)

// UFFD_EVENT_PAGEFAULT flags
const (
	UFFD_PAGEFAULT_FLAG_WRITE = 1 << iota // 1 << 0
	UFFD_PAGEFAULT_FLAG_WP                // 1 << 1
	UFFD_PAGEFAULT_FLAG_MINOR             // 1 << 2
)

// UFFDIO_CONTINUE(2) ioctl mode
const (
	UFFDIO_CONTINUE_MODE_DONTWAKE = 1 << iota // 1 << 0
	UFFDIO_CONTINUE_MODE_WP                   // 1 << 1
)

// UFFDIO_COPY(2) ioctl mode
const (
	UFFDIO_COPY_MODE_DONTWAKE = 1 << iota // 1 << 0
	UFFDIO_COPY_MODE_WP                   // 1 << 1
)

// UFFDIO_MOVE(2) ioctl mode
const (
	UFFDIO_MOVE_MODE_DONTWAKE        = 1 << iota // 1 << 0
	UFFDIO_MOVE_MODE_ALLOW_SRC_HOLES             // 1 << 1
)

// UFFDIO_POISON(2) ioctl mode
const (
	UFFDIO_POISON_MODE_DONTWAKE = 1 << iota // 1 << 0
)

// UFFDIO_REGISTER(2) ioctl mode
const (
	UFFDIO_REGISTER_MODE_MISSING = 1 << iota // 1 << 0
	UFFDIO_REGISTER_MODE_WP                  // 1 << 1
	UFFDIO_REGISTER_MODE_MINOR               // 1 << 2
)

// UFFDIO_WRITEPROTECT(2) ioctl mode
const (
	UFFDIO_WRITEPROTECT_MODE_WP       = 1 << iota // 1 << 0
	UFFDIO_WRITEPROTECT_MODE_DONTWAKE             // 1 << 1
)

// UFFDIO_ZEROPAGE(2) ioctl mode
const (
	UFFDIO_ZEROPAGE_MODE_DONTWAKE = 1 << iota // 1 << 0
)
