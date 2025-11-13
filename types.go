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
	Start uint64
	Len   uint64
}

// UffdioRegister is used with UFFDIO_REGISTER.
type UffdioRegister struct {
	Range  UffdioRange
	Mode   uint64
	Ioctls uint64
}

// UffdioCopy is used with UFFDIO_COPY.
type UffdioCopy struct {
	Dst  uint64
	Src  uint64
	Len  uint64
	Mode uint64
	Copy int64
}

// UffdioZeropage is used with UFFDIO_ZEROPAGE.
type UffdioZeropage struct {
	Range    UffdioRange
	Mode     uint64
	Zeropage int64
}

// UffdioWriteprotect is used with UFFDIO_WRITEPROTECT.
type UffdioWriteprotect struct {
	Range UffdioRange
	Mode  uint64
}

// UffdioContinue is used with UFFDIO_CONTINUE.
type UffdioContinue struct {
	Range  UffdioRange
	Mode   uint64
	Mapped int64
}

// UffdioPoison is used with UFFDIO_POISON.
type UffdioPoison struct {
	Range   UffdioRange
	Mode    uint64
	Updated int64
}

// UffdioMove is used with UFFDIO_MOVE.
type UffdioMove struct {
	Dst  uint64
	Src  uint64
	Len  uint64
	Mode uint64
	Move int64
}

type UffdMsg struct {
	Event uint8
	_     [7]byte // padding

	Data [24]byte
}

type UffdMsgPagefault struct {
	Flags   uint64
	Address uint64
	Ptid    uint32
	_       uint32 // padding
}

func (m *UffdMsg) GetPagefault() *UffdMsgPagefault {
	return (*UffdMsgPagefault)(unsafe.Pointer(&m.Data[0]))
}

type UffdMsgFork struct {
	Ufd uint32
}

func (m *UffdMsg) GetFork() *UffdMsgFork {
	return (*UffdMsgFork)(unsafe.Pointer(&m.Data[0]))
}

type UffdMsgRemap struct {
	From uint64
	To   uint64
	Len  uint64
}

func (m *UffdMsg) GetRemap() *UffdMsgRemap {
	return (*UffdMsgRemap)(unsafe.Pointer(&m.Data[0]))
}

type UffdMsgRemove struct {
	Start uint64
	End   uint64
}

func (m *UffdMsg) GetRemove() *UffdMsgRemove {
	return (*UffdMsgRemove)(unsafe.Pointer(&m.Data[0]))
}
