/* SPDX-License-Identifier: BSD-2-Clause */

package userfaultfd

import (
	"unsafe"
)

// UffdioApi is used with UFFDIO_API.
type UffdioApi struct {
	Api      uint64
	Features uint64
	Ioctls   uint64
}

// UffdioRange is used with UFFDIO_UNREGISTER, UFFDIO_WAKE, etc.
type UffdioRange struct {
	Start uint64 // Start of range
	Len   uint64 // Size of range (bytes)
}

// UffdioRegister is used with UFFDIO_REGISTER.
type UffdioRegister struct {
	Range  UffdioRange
	Mode   uint64 // Desired mode of operation: UFFDIO_REGISTER_MODE_MISSING, UFFDIO_REGISTER_MODE_WP & UFFDIO_REGISTER_MODE_MINOR
	Ioctls uint64 // Returns: Available ioctl()s
}

// UffdioCopy is used with UFFDIO_COPY.
type UffdioCopy struct {
	Dst  uint64 // Destination of copy
	Src  uint64 // Source of copy
	Len  uint64 // Number of bytes to copy
	Mode uint64 // Flags controlling behavior of copy: UFFDIO_COPY_MODE_DONTWAKE & UFFDIO_COPY_MODE_WP
	Copy int64  // Returns Number of bytes copied, or negated error
}

// UffdioZeropage is used with UFFDIO_ZEROPAGE.
type UffdioZeropage struct {
	Range    UffdioRange
	Mode     uint64 // Flags controlling behavior: UFFDIO_ZEROPAGE_MODE_DONTWAKE
	Zeropage int64  // Returns: Number of bytes zeroed
}

// UffdioWriteprotect is used with UFFDIO_WRITEPROTECT.
type UffdioWriteprotect struct {
	Range UffdioRange // Range to change write permission
	Mode  uint64      // Mode to change write permission: UFFDIO_WRITEPROTECT_MODE_WP & UFFDIO_WRITEPROTECT_MODE_DONTWAKE
}

// UffdioContinue is used with UFFDIO_CONTINUE.
type UffdioContinue struct {
	Range  UffdioRange // Range to install PTEs for and continue
	Mode   uint64      // Flags controlling the behavior of continue: UFFDIO_CONTINUE_MODE_DONTWAKE
	Mapped int64       // Returns: Number of bytes mapped, or negated error
}

// UffdioPoison is used with UFFDIO_POISON.
type UffdioPoison struct {
	Range   UffdioRange // Range to install poison PTE markers in
	Mode    uint64      // Flags controlling the behavior of poison: UFFDIO_POISON_MODE_DONTWAKE
	Updated int64       // Returns: Number of bytes poisoned, or negated error
}

// UffdioMove is used with UFFDIO_MOVE.
type UffdioMove struct {
	Dst  uint64 // Destination of move
	Src  uint64 // Source of move
	Len  uint64 // Number of bytes to move
	Mode uint64 // Flags controlling behavior of move
	Move int64  // Returns: Number of bytes moved, or negated error
}

type UffdMsg struct {
	Event uint8
	_     [7]byte // padding
	Data  [24]byte
}

type UffdMsgPagefault struct {
	Flags   uint64 // Flags describing fault
	Address uint64 // Faulting address. Needs UFFD_FEATURE_EXACT_ADDRESS
	Ptid    uint32 // Thread ID of the fault. Needs UFFD_FEATURE_THREAD_ID
	_       uint32 // padding
}

func (m *UffdMsg) GetPagefault() *UffdMsgPagefault {
	return (*UffdMsgPagefault)(unsafe.Pointer(&m.Data[0]))
}

type UffdMsgFork struct {
	Ufd uint32 // Userfault file descriptor of the child process
}

func (m *UffdMsg) GetFork() *UffdMsgFork {
	return (*UffdMsgFork)(unsafe.Pointer(&m.Data[0]))
}

type UffdMsgRemap struct {
	From uint64 // Old address of remapped area
	To   uint64 // New address of remapped area
	Len  uint64 // Original mapping size
}

func (m *UffdMsg) GetRemap() *UffdMsgRemap {
	return (*UffdMsgRemap)(unsafe.Pointer(&m.Data[0]))
}

type UffdMsgRemove struct {
	Start uint64 // Start address of removed area
	End   uint64 // End address of removed area
}

func (m *UffdMsg) GetRemove() *UffdMsgRemove {
	return (*UffdMsgRemove)(unsafe.Pointer(&m.Data[0]))
}
