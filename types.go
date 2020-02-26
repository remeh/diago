package main

import "fmt"

type StringsMap map[uint64]string
type LocationsMap map[uint64]Location
type FunctionsMap map[uint64]Function

type Sample struct {
	Functions    []Function
	Value        int64
	PercentTotal float64
}

type Samples []Sample

type Location struct {
	Function
}

type Function struct {
	Name       string
	File       string
	LineNumber uint64
}

func (f Function) String(lineNumber bool) string {
	if lineNumber {
		return fmt.Sprintf("%s %s:%d", f.Name, f.File, f.LineNumber)
	}
	return fmt.Sprintf("%s %s", f.Name, f.File)
}
