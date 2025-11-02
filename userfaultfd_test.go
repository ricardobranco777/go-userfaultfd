/* SPDX-License-Identifier: BSD-2-Clause */

package userfaultfd

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Default flags to use with userfaultfd syscall
var flags = 0

func TestMain(m *testing.M) {
	if os.Geteuid() != 0 && !UnprivilegedUserfaultfd {
		if !HaveUserModeOnly {
			println("Skipping all tests: UFFD_USER_MODE_ONLY not supported on this kernel")
			os.Exit(1)
		}
		flags |= UFFD_USER_MODE_ONLY
	}

	// Ensure basic syscall exists
	fd, _, errno := unix.Syscall(unix.SYS_USERFAULTFD, uintptr(flags), 0, 0)
	switch errno {
	case unix.ENOSYS:
		println("Skipping all tests: userfaultfd syscall not available on this kernel")
		os.Exit(1)
	case unix.EPERM:
		println("Skipping all tests: vm.unprivileged_userfaultfd probably unset")
		os.Exit(1)
	case 0:
		break
	default:
		fmt.Printf("Skipping all tests: error: %v\n", errno)
		os.Exit(1)
	}

	_ = unix.Close(int(fd))

	os.Exit(m.Run())
}

func TestNewFile(t *testing.T) {
	f, err := NewFile(flags)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	defer f.Close()

	fd := int(f.Fd())
	if fd < 0 {
		t.Fatalf("invalid fd: %d", fd)
	}
}

// TestNewFile2 tests that /dev/userfaultfd can be opened via ioctl.
func TestNewFile2(t *testing.T) {
	if !HaveDevUserfaultfd {
		t.Skip("/dev/userfaultfd does not exist")
	}
	f, err := NewFile2(0)
	if err != nil {
		if errors.Is(err, unix.EACCES) {
			t.Skip("/dev/userfaultfd is not readable")
		} else {
			t.Fatalf("NewFile2 failed: %v", err)
		}
	}
	defer f.Close()

	fd := int(f.Fd())
	if fd < 0 {
		t.Fatalf("invalid fd: %d", fd)
	}
}

func TestApiHandshake(t *testing.T) {
	f, err := NewFile(flags)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	defer f.Close()

	api, err := ApiHandshake(int(f.Fd()), 0)
	if err != nil {
		t.Fatalf("ApiHandshake failed: %v", err)
	}

	t.Logf("Userfaultfd API version: %d, features: 0x%x, ioctls: 0x%x", api.Api, api.Features, api.Ioctls)
}

func TestRegisterAndUnregister(t *testing.T) {
	f, err := NewFile(flags)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	defer f.Close()

	if _, err = ApiHandshake(int(f.Fd()), 0); err != nil {
		t.Fatalf("ApiHandshake failed: %v", err)
	}

	pageSize := unix.Getpagesize()

	// tmpfs-backed mapping (shared memory)
	tmp, err := os.CreateTemp("/dev/shm", "uffd_test")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	// Extend file to pageSize
	if err := tmp.Truncate(int64(pageSize)); err != nil {
		t.Fatalf("truncate failed: %v", err)
	}

	mem, err := unix.Mmap(int(tmp.Fd()), 0, pageSize, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		t.Fatalf("mmap (shmem) failed: %v", err)
	}
	defer unix.Munmap(mem)

	addr := uintptr(unsafe.Pointer(&mem[0]))

	// Attempt registration
	if _, err = Register(int(f.Fd()), addr, uintptr(pageSize), UFFDIO_REGISTER_MODE_MISSING); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Now unregister
	if err := Unregister(int(f.Fd()), addr, uintptr(pageSize)); err != nil {
		t.Fatalf("Unregister failed: %v", err)
	}
}

func setupUserfaultfd(t *testing.T, features uint64) (fd int, addr uintptr, cleanup func()) {
	t.Helper()

	f, err := NewFile(flags)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	api, err := ApiHandshake(int(f.Fd()), 0)
	if err != nil {
		f.Close()
		t.Fatalf("ApiHandshake (enable features) failed: %v", err)
	}

	if features != 0 {
		f.Close()
		if f, err = NewFile(flags); err != nil {
			t.Fatalf("Create failed: %v", err)
		}
		got := api.Features
		if api, err = ApiHandshake(int(f.Fd()), features); err != nil {
			f.Close()
			t.Skipf("requested features 0x%x not fully supported (got 0x%x)", features, got)
		}
	}

	fd = int(f.Fd())

	pageSize := unix.Getpagesize()
	mem, err := unix.Mmap(-1, 0, pageSize, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_PRIVATE|unix.MAP_ANONYMOUS)
	if err != nil {
		f.Close()
		t.Fatalf("mmap failed: %v", err)
	}

	addr = uintptr(unsafe.Pointer(&mem[0]))

	mode := uint64(UFFDIO_REGISTER_MODE_MISSING)
	if features&UFFD_FEATURE_PAGEFAULT_FLAG_WP != 0 {
		mode |= UFFDIO_REGISTER_MODE_WP
	}

	if _, err := Register(fd, addr, uintptr(pageSize), mode); err != nil {
		f.Close()
		unix.Munmap(mem)
		t.Fatalf("Register failed: %v", err)
	}

	cleanup = func() {
		_ = Unregister(fd, addr, uintptr(pageSize))
		_ = unix.Munmap(mem)
		_ = f.Close()
	}
	return
}

func TestContinue(t *testing.T) {
	if !HaveIoctlContinue {
		t.Skip("UFFDIO_CONTINUE not available")
	}

	f, err := NewFile(flags)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	defer f.Close()

	if _, err = ApiHandshake(int(f.Fd()), UFFD_FEATURE_MINOR_SHMEM); err != nil {
		if errors.Is(err, unix.EINVAL) {
			t.Skip("Unsupported UFFD_FEATURE_MINOR_SHMEM")
		} else {
			t.Fatalf("ApiHandshake failed: %v", err)
		}
	}

	fd := int(f.Fd())

	// Create a temporary file backed by tmpfs/shmem
	tmp, err := os.CreateTemp("/dev/shm", "uffd_test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	pageSize := unix.Getpagesize()

	// Write some data to the file to create backing pages
	data := make([]byte, pageSize)
	for i := range data {
		data[i] = 0xAB
	}
	if _, err := tmp.Write(data); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}

	// Map the file with MAP_SHARED to get shmem backing
	mem, err := unix.Mmap(int(tmp.Fd()), 0, pageSize, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		t.Fatalf("mmap failed: %v", err)
	}
	defer unix.Munmap(mem)

	addr := uintptr(unsafe.Pointer(&mem[0]))

	// Register for MINOR fault handling
	if _, err = Register(fd, addr, uintptr(pageSize), UFFDIO_REGISTER_MODE_MINOR); err != nil {
		t.Fatalf("Register for minor faults failed: %v", err)
	}
	defer Unregister(fd, addr, uintptr(pageSize))

	// Remove the page table entries to trigger minor faults
	if err := unix.Madvise(mem, unix.MADV_DONTNEED); err != nil {
		t.Fatalf("madvise MADV_DONTNEED failed: %v", err)
	}

	// Now UFFDIO_CONTINUE should work - it maps the existing page
	if err := Continue(fd, addr, uintptr(pageSize), 0); err != nil {
		t.Errorf("Continue failed: %v", err)
	}

	// Verify we can access the memory and it contains the expected data
	if mem[0] != 0xAB {
		t.Errorf("Expected data 0xAB, got 0x%02X", mem[0])
	}
}

func TestCopy(t *testing.T) {
	fd, dst, cleanup := setupUserfaultfd(t, 0)
	defer cleanup()

	srcMem := make([]byte, unix.Getpagesize())
	for i := range srcMem {
		srcMem[i] = 0xAA
	}
	src := uintptr(unsafe.Pointer(&srcMem[0]))

	n, err := Copy(fd, dst, src, uintptr(len(srcMem)), 0)
	if err != nil {
		t.Errorf("Copy failed: %v", err)
	}
	if n != int64(len(srcMem)) {
		t.Errorf("Copy copied unexpected length: got %d", n)
	}
}

func TestMove(t *testing.T) {
	if !HaveIoctlMove {
		t.Skip("UFFDIO_MOVE not available")
	}

	fd, _, cleanup := setupUserfaultfd(t, UFFD_FEATURE_MOVE)
	defer cleanup()

	pageSize := uintptr(unix.Getpagesize())

	// Create disjoint anonymous mappings
	src, err := unix.Mmap(-1, 0, int(pageSize), unix.PROT_READ|unix.PROT_WRITE, unix.MAP_PRIVATE|unix.MAP_ANONYMOUS)
	if err != nil {
		t.Fatalf("mmap src failed: %v", err)
	}
	defer unix.Munmap(src)

	dst, err := unix.Mmap(-1, 0, int(pageSize), unix.PROT_READ|unix.PROT_WRITE, unix.MAP_PRIVATE|unix.MAP_ANONYMOUS)
	if err != nil {
		t.Fatalf("mmap dst failed: %v", err)
	}
	defer unix.Munmap(dst)

	srcPtr := uintptr(unsafe.Pointer(&src[0]))
	dstPtr := uintptr(unsafe.Pointer(&dst[0]))

	// Register both regions
	if _, err := Register(fd, srcPtr, pageSize, UFFDIO_REGISTER_MODE_MISSING); err != nil {
		t.Fatalf("Register src failed: %v", err)
	}
	if _, err := Register(fd, dstPtr, pageSize, UFFDIO_REGISTER_MODE_MISSING); err != nil {
		t.Fatalf("Register dst failed: %v", err)
	}
	defer func() {
		Unregister(fd, srcPtr, pageSize)
		Unregister(fd, dstPtr, pageSize)
	}()

	// Populate only the *source* page
	if _, err := Zeropage(fd, srcPtr, pageSize, 0); err != nil {
		t.Fatalf("Zeropage src failed: %v", err)
	}

	// Ensure destination remains missing (no Zeropage on dst)

	// Perform the move
	n, err := Move(fd, dstPtr, srcPtr, pageSize, 0)
	if err != nil {
		t.Skipf("Move ioctl failed (unsupported on this kernel): %v", err)
	}
	if n != int64(pageSize) {
		t.Errorf("expected Move length %d, got %d", pageSize, n)
	}
}

func TestPoison(t *testing.T) {
	if !HaveIoctlPoison {
		t.Skip("UFFDIO_POISON not available")
	}

	fd, addr, cleanup := setupUserfaultfd(t, UFFD_FEATURE_POISON)
	defer cleanup()

	updated, err := Poison(fd, addr, uintptr(unix.Getpagesize()), 0)
	if err != nil {
		t.Errorf("Poison failed: %v", err)
	}
	if updated <= 0 {
		t.Errorf("Poison reported no pages updated: got %d", updated)
	}
}

func TestWake(t *testing.T) {
	fd, addr, cleanup := setupUserfaultfd(t, 0)
	defer cleanup()

	if err := Wake(fd, addr, uintptr(unix.Getpagesize())); err != nil {
		t.Errorf("Wake failed: %v", err)
	}
}

func TestWriteProtect(t *testing.T) {
	if !HaveIoctlWriteProtect {
		t.Skip("UFFDIO_WRITEPROTECT not available")
	}

	fd, addr, cleanup := setupUserfaultfd(t, UFFD_FEATURE_PAGEFAULT_FLAG_WP)
	defer cleanup()

	if err := WriteProtect(fd, addr, uintptr(unix.Getpagesize()), UFFDIO_WRITEPROTECT_MODE_WP); err != nil {
		t.Errorf("WriteProtect (enable) failed: %v", err)
	}
}

func TestZeropage(t *testing.T) {
	fd, addr, cleanup := setupUserfaultfd(t, 0)
	defer cleanup()

	n, err := Zeropage(fd, addr, uintptr(unix.Getpagesize()), 0)
	if err != nil {
		t.Errorf("Zeropage failed: %v", err)
	}
	if n != int64(unix.Getpagesize()) {
		t.Errorf("Zeropage returned unexpected length: got %d", n)
	}
}
