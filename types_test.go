/* SPDX-License-Identifier: BSD-2-Clause */

package userfaultfd

import (
	"testing"
	"unsafe"
)

func TestStructSizes(t *testing.T) {
	tests := []struct {
		name string
		got  uintptr
		want uintptr
	}{
		{"UffdMsg", unsafe.Sizeof(UffdMsg{}), 32},
		{"UffdioApi", unsafe.Sizeof(UffdioApi{}), 24},
		{"UffdioContinue", unsafe.Sizeof(UffdioContinue{}), 32},
		{"UffdioCopy", unsafe.Sizeof(UffdioCopy{}), 40},
		{"UffdioMove", unsafe.Sizeof(UffdioMove{}), 40},
		{"UffdioPoison", unsafe.Sizeof(UffdioPoison{}), 32},
		{"UffdioRange", unsafe.Sizeof(UffdioRange{}), 16},
		{"UffdioRegister", unsafe.Sizeof(UffdioRegister{}), 32},
		{"UffdioWriteprotect", unsafe.Sizeof(UffdioWriteprotect{}), 24},
		{"UffdioZeropage", unsafe.Sizeof(UffdioZeropage{}), 32},
		{"UffdMsgPagefault", unsafe.Sizeof(UffdMsgPagefault{}), 24},
	}

	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s size = %d, want %d", tt.name, tt.got, tt.want)
		}
	}
}
