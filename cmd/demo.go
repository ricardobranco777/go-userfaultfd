// SPDX-License-Identifier: BSD-2-Clause

// Go translation of the example in userfaultfd(2)

package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"

	userfaultfd "github.com/ricardobranco777/go-userfaultfd"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s pages\n", os.Args[0])
		os.Exit(1)
	}

	pageSize := unix.Getpagesize()

	pages, err := strconv.ParseUint(os.Args[1], 10, 64)
	if err != nil {
		log.Fatalf("invalid number: %v", err)
	}
	size := int(pages) * pageSize

	mem, err := unix.Mmap(-1, 0, size, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_PRIVATE|unix.MAP_ANONYMOUS)
	if err != nil {
		log.Fatalf("mmap failed: %v", err)
	}
	defer unix.Munmap(mem)

	addr := uintptr(unsafe.Pointer(&mem[0]))
	fmt.Printf("Address returned by mmap() = %#x\n", addr)

	u, err := userfaultfd.New(userfaultfd.UFFD_USER_MODE_ONLY|unix.O_NONBLOCK, 0)
	if err != nil {
		log.Fatalf("userfaultfd New failed: %v", err)
	}
	defer u.Close()

	if _, err := u.Register(addr, size, userfaultfd.UFFDIO_REGISTER_MODE_MISSING); err != nil {
		log.Fatalf("UFFDIO_REGISTER failed: %v", err)
	}
	defer u.Unregister(addr, size)

	go faultHandler(u, pageSize)

	// Touch memory at 1024-byte intervals
	i := 0xF // ensure not page-aligned
	for i < size {
		c := mem[i] // triggers page fault the first time for each page
		fmt.Printf("Read address %#x in main(): %q\n", addr+uintptr(i), c)
		i += 1024
		time.Sleep(100 * time.Millisecond)
	}
}

func faultHandler(u *userfaultfd.Uffd, pageSize int) {
	page, err := unix.Mmap(-1, 0, pageSize, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_PRIVATE|unix.MAP_ANONYMOUS)
	if err != nil {
		log.Fatalf("handler: mmap staging page failed: %v", err)
	}
	defer unix.Munmap(page)

	ch := byte('A')

	for {
		msg, err := u.ReadMsg()
		if err != nil {
			log.Fatalf("handler: ReadMsg failed: %v", err)
		}
		if msg.Event != userfaultfd.UFFD_EVENT_PAGEFAULT {
			continue
		}

		pf := msg.GetPagefault()
		addr := uintptr(pf.Address)

		fmt.Printf("\nfault_handler_thread():\n")
		fmt.Printf("    UFFD_EVENT_PAGEFAULT event: flags = %x; address = %x\n", pf.Flags, pf.Address)

		for i := 0; i < pageSize; i++ {
			page[i] = ch
		}
		ch++

		// u.Copy would also work
		var n int64
		if userfaultfd.HaveIoctlMove {
			n, err = u.Move(addr, uintptr(unsafe.Pointer(&page[0])), pageSize, 0)
		} else {
			n, err = u.Copy(addr, uintptr(unsafe.Pointer(&page[0])), pageSize, 0)
		}
		if err != nil {
			log.Fatalf("handler: UFFDIO_COPY failed at 0x%x: %v", addr, err)
		}
		fmt.Printf("\t(uffdio_copy.copy returned %d)\n", n)
	}
}
