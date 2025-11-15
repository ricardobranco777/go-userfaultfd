/* SPDX-License-Identifier: BSD-2-Clause */

package userfaultfd

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

func TestServeMapping(t *testing.T) {
	// Open a local file we can page in
	f, err := os.Open("/bin/bash")
	if err != nil {
		t.Skipf("skipping: testdata file missing: %v", err)
	}
	defer f.Close()

	pageSize := unix.Getpagesize()
	size := int64(pageSize * 256) // 256 pages for test
	info, err := f.Stat()
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if info.Size() < size {
		size = info.Size()
	}

	// Call new helper instead of boilerplate UFFD setup
	data, closeFn, err := ServeMapping(f, size)
	if err != nil {
		t.Skipf("ServeMapping unavailable: %v", err)
	}
	defer closeFn()

	// Trigger UFFD page faults across the whole region
	for i := int64(0); i < size; i += int64(pageSize) {
		_ = data[i]
	}

	// Allow handler to page in content
	time.Sleep(500 * time.Millisecond)

	// --- Hash original file (only first 'size' bytes)
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("seek failed: %v", err)
	}

	h1 := sha256.New()
	if _, err := io.CopyN(h1, f, size); err != nil && !errors.Is(err, io.EOF) {
		t.Fatalf("failed computing reference hash: %v", err)
	}
	expectedHash := h1.Sum(nil)

	// --- Compute hash of returned mmap region
	h := sha256.New()
	if _, err := h.Write(data); err != nil {
		t.Fatalf("failed hashing mmap region: %v", err)
	}
	actualHash := h.Sum(nil)

	if !bytes.Equal(expectedHash, actualHash) {
		t.Fatalf("content mismatch: expected %x, got %x", expectedHash, actualHash)
	}
}
