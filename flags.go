package main

import "flag"

type Config struct {
	File string
}

var config Config

func init() {
	flag.StringVar(&config.File, "file", "", "Profile or heap snapshot file to read")
	flag.Parse()
}
