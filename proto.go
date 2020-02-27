package main

import (
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/gogo/protobuf/proto"
	"github.com/remeh/diago/pprof"
)

func readProtoFile(filename string) (*pprof.Profile, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("readProtoFile: os.Open: %v", err)
	}

	g, err := gzip.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("readProtoFile: gzip.NewReader: %v", err)
	}

	data, err := ioutil.ReadAll(g)
	if err != nil {
		return nil, fmt.Errorf("readProtoFile: ioutil.ReadAll: %v", err)
	}

	var profile pprof.Profile
	if err := proto.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("readProtoFile: proto.Unmarshal: %v", err)
	}

	return &profile, nil
}
