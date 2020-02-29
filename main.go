package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"

	"github.com/remeh/diago/pprof"
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

	if pprofProfile, err = readProtoFile(config.File); err != nil {
		fmt.Println("err:", err)
		os.Exit(-1)
	}

	// start the gui
	// ----------------------

	gui := NewGUI(pprofProfile)
	gui.OpenWindow()
}
