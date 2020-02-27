package main

import (
	"fmt"
	"log"
	"time"

	"github.com/remeh/diago/pprof"
)

type Profile struct {
	Samples
	SamplingDuration time.Duration
	CaptureDuration  time.Duration

	functionsMap           FunctionsMap
	functionsMapByLocation FunctionsMap
	locationsMap           LocationsMap
	stringsMap             StringsMap
}

func NewProfile(p *pprof.Profile) (*Profile, error) {
	// start by building some maps because everything
	// is indexed in various maps.
	// ----------------------

	// strings map
	stringsMap := buildStringsTable(p)

	// functions map
	functionsMap := buildFunctionsMap(p, stringsMap)

	// locations map
	locationsMap, functionsMapByLocation := buildLocationsMap(p, stringsMap, functionsMap)

	// let's now build the profile
	// ----------------------

	if stringsMap[uint64(p.GetPeriodType().Type)] != "cpu" {
		return nil, fmt.Errorf("Diago only supports cpu profile for now")
	}

	var samples Samples

	for _, pprofSample := range p.Sample {
		var sample Sample
		for i := len(pprofSample.LocationId) - 1; i >= 0; i-- {
			l := pprofSample.LocationId[i]
			sample.Functions = append(sample.Functions, functionsMapByLocation[l])
			// TODO(remy): isn't there anything to do with the [0]?
			//			   I think it represents how many samples were aggregated
			//			   into this one, meaning we don't really need to use [0]
			sample.Value = pprofSample.GetValue()[1]
		}
		samples = append(samples, sample)
	}

	// compute the total sampling time
	var totalSum int64
	for _, s := range samples {
		totalSum += s.Value
	}

	// compute the percentage for every sample
	for i, s := range samples {
		s.PercentTotal = float64(s.Value) / (float64(totalSum)) * 100.0
		samples[i] = s
	}

	return &Profile{
		Samples:          samples,
		SamplingDuration: time.Duration(totalSum),
		CaptureDuration:  time.Duration(p.GetDurationNanos()),

		functionsMap:           functionsMap,
		functionsMapByLocation: functionsMapByLocation,
		locationsMap:           locationsMap,
		stringsMap:             stringsMap,
	}, nil
}

func (p *Profile) BuildTree(treeName string, aggregateByFunction bool, searchField string) *FunctionsTree {
	// prepare the tree
	tree := NewFunctionsTree(treeName)

	// fill the tree
	for _, s := range p.Samples {
		node := tree.root
		for _, f := range s.Functions {
			node = node.AddFunction(f, s.Value, s.PercentTotal, aggregateByFunction)
		}
	}

	if tree.root != nil {
		tree.root.filter(searchField)
	}

	tree.sort()

	return tree
}

func buildLocationsMap(profile *pprof.Profile, stringsMap StringsMap, functionsMap FunctionsMap) (LocationsMap, FunctionsMap) {
	rv := make(LocationsMap)
	lrv := make(FunctionsMap)

	for _, location := range profile.Location {
		// TODO(remy): there could be many lines here.
		// > A location has multiple lines if it reflects multiple program sources,
		// > for example if representing inlined call stacks.
		if len(location.Line) > 1 {
			log.Println("warn: many lines in a Location, unsupported for now")
		}

		if location.Line[0] == nil {
			continue
		}

		line := location.Line[0]
		loc := Location{
			Function: functionsMap[line.GetFunctionId()],
		}
		loc.Function.LineNumber = uint64(line.GetLine())
		rv[uint64(location.GetId())] = loc

		// TODO(remy): because there could be many lines, there could be
		//			   many functions here. Didn't see this in any profile
		//			   created by Go http/pprof though.

		// set the line number in functions
		f := functionsMap[line.GetFunctionId()]
		f.LineNumber = uint64(line.GetLine())

		lrv[uint64(location.GetId())] = f
		functionsMap[line.GetFunctionId()] = f
	}

	return rv, lrv
}

func buildFunctionsMap(profile *pprof.Profile, stringsMap StringsMap) FunctionsMap {
	rv := make(FunctionsMap)
	for _, f := range profile.Function {
		rv[f.GetId()] = Function{
			Name: stringsMap[uint64(f.GetName())],
			File: stringsMap[uint64(f.GetFilename())],
		}
	}
	return rv
}

func buildStringsTable(profile *pprof.Profile) StringsMap {
	rv := make(StringsMap)
	for i, v := range profile.GetStringTable() {
		rv[uint64(i)] = v
	}
	return rv
}
