// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	rf "rsc.io/rf"
	"rsc.io/rf/refactor"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: rf [-diff] script\n")
	os.Exit(2)
}

func main() {
	log.SetPrefix("rf: ")
	log.SetFlags(0)

	var (
		showDiff bool
		allPlat  bool
	)
	flag.BoolVar(&showDiff, "diff", false, "show diff instead of writing files")
	flag.BoolVar(&allPlat, "allplat", false, "apply refactoring across all GOOS/GOARCH combinations")
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		usage()
	}
	script := args[0]

	r, err := refactor.New(".")
	if err != nil {
		log.Fatal(err)
	}
	r.ShowDiff = showDiff

	if allPlat {
		r.Configs, err = refactor.ConfigsAllPlatforms()
		if err != nil {
			log.Fatal(err)
		}
	}

	if err := rf.Run(r, script); err != nil {
		log.Fatal(err)
	}
}
