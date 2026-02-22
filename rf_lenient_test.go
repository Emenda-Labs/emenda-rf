// Copyright 2024 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rf

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"rsc.io/rf/refactor"
)

// TestLenientMvCrossPackage tests that "mv pkg.OldFunc pkg.NewFunc" renames
// both the definition in the subpackage and all call sites across packages,
// even when Lenient mode is enabled.
func TestLenientMvCrossPackage(t *testing.T) {
	dir := t.TempDir()

	// Write go.mod
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module testmod\ngo 1.21\n"), 0666); err != nil {
		t.Fatal(err)
	}

	// Write pkg/pkg.go - defines OldFunc
	if err := os.MkdirAll(filepath.Join(dir, "pkg"), 0777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "pkg", "pkg.go"), []byte("package pkg\n\nfunc OldFunc() {}\n"), 0666); err != nil {
		t.Fatal(err)
	}

	// Write main.go - calls pkg.OldFunc
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte(`package main

import "testmod/pkg"

func main() { pkg.OldFunc() }
`), 0666); err != nil {
		t.Fatal(err)
	}

	r, err := refactor.New(dir)
	if err != nil {
		t.Fatal("refactor.New:", err)
	}
	r.Lenient = true

	var stdout, stderr bytes.Buffer
	r.Stdout = &stdout
	r.Stderr = &stderr

	if err := Run(r, "mv pkg.OldFunc pkg.NewFunc\n"); err != nil {
		t.Logf("stderr: %s", stderr.String())
		t.Fatal("Run error:", err)
	}

	t.Logf("stdout:\n%s", stdout.String())
	t.Logf("stderr:\n%s", stderr.String())

	// Verify the diff contains renames in both files.
	out := stdout.String()
	if !bytes.Contains([]byte(out), []byte("OldFunc")) || !bytes.Contains([]byte(out), []byte("NewFunc")) {
		// If no diff output at all, check whether files were written.
		t.Log("No diff output detected, checking written files...")
	}

	// Check that pkg/pkg.go was modified (OldFunc -> NewFunc in definition).
	pkgFile, err := os.ReadFile(filepath.Join(dir, "pkg", "pkg.go"))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("pkg/pkg.go after rename:\n%s", string(pkgFile))
	if bytes.Contains(pkgFile, []byte("OldFunc")) {
		t.Error("pkg/pkg.go still contains OldFunc - definition was NOT renamed")
	}
	if !bytes.Contains(pkgFile, []byte("NewFunc")) {
		t.Error("pkg/pkg.go does not contain NewFunc - definition rename failed")
	}

	// Check that main.go was modified (pkg.OldFunc() -> pkg.NewFunc()).
	mainFile, err := os.ReadFile(filepath.Join(dir, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("main.go after rename:\n%s", string(mainFile))
	if bytes.Contains(mainFile, []byte("OldFunc")) {
		t.Error("main.go still contains OldFunc - call site was NOT renamed")
	}
	if !bytes.Contains(mainFile, []byte("NewFunc")) {
		t.Error("main.go does not contain NewFunc - call site rename failed")
	}
}
