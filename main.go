package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"remy.io/diago/pprof"
)

func main() {
	runtime.LockOSThread()
	if config.File == "" {
		flag.PrintDefaults()
		os.Exit(-1)
	}

	var err error

	// read the pprof file
	// ----------------------

	var pprofProfile *pprof.Profile
	var profile *Profile

	if pprofProfile, err = readProtoFile(config.File); err != nil {
		fmt.Println("err:", err)
		os.Exit(-1)
	}

	// convert to a profile diago object
	// ----------------------

	if profile, err = NewProfile(pprofProfile); err != nil {
		fmt.Println("err:", err)
		os.Exit(-1)
	}

	// start the gui
	// ----------------------

	gui := NewGUI(profile, profile.BuildTree(config.File, true, ""))
	gui.OpenWindow()
}
