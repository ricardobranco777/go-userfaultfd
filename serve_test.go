/* SPDX-License-Identifier: BSD-2-Clause */

package userfaultfd

import (
	"bytes"
	"crypto/sha256"
	"io"
	"os"
	"testing"
)

func TestServe(t *testing.T) {
	f, err := os.Open("/bin/bash")
	if err != nil {
		t.Skipf("skipping: testdata file missing: %v", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	size := int(info.Size())

	data, cleanup, err := Serve(f, size)
	if err != nil {
		t.Skipf("Serve unavailable: %v", err)
	}
	defer cleanup()

	if _, err := f.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("seek failed: %v", err)
	}
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		t.Fatalf("failed computing reference hash: %v", err)
	}
	expectedHash := h.Sum(nil)

	actualHash := sha256.Sum256(data[:size])

	if !bytes.Equal(expectedHash[:], actualHash[:]) {
		t.Fatalf("content mismatch: expected %x, got %x", expectedHash, actualHash)
	}
}
