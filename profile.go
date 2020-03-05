package main

import (
	"fmt"
	"os"
	"time"

	"github.com/remeh/diago/pprof"
)

type Profile struct {
	Samples
	TotalSampling   uint64
	CaptureDuration time.Duration

	// "cpu" or "heap"
	Type string

	functionsMapByLocation ManyFunctionsMap
	locationsMap           LocationsMap
	stringsMap             StringsMap
}

func NewProfile(p *pprof.Profile, mode sampleMode) (*Profile, error) {
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

	typ := ReadProfileType(p)

	if typ != "cpu" && typ != "space" {
		return nil, fmt.Errorf("unsupported type: %s", typ)
	}

	profile := readProfile(p, stringsMap, functionsMapByLocation, locationsMap, mode)

	switch typ {
	case "cpu":
		profile.Type = "cpu"
	case "space":
		profile.Type = "heap"
	}

	return profile, nil
}

func ReadProfileType(p *pprof.Profile) string {
	return p.StringTable[uint64(p.GetPeriodType().Type)]
}

func readProfile(p *pprof.Profile, stringsMap StringsMap, functionsMapByLocation ManyFunctionsMap,
	locationsMap LocationsMap, mode sampleMode) *Profile {

	var samples Samples
	var idx int

	switch {
	case mode == ModeDefault:
		fallthrough
	case ReadProfileType(p) == "cpu" && mode == ModeCpu:
		idx = 1
	case ReadProfileType(p) == "space" && mode == ModeHeapAlloc:
		idx = 1
	case ReadProfileType(p) == "space" && mode == ModeHeapInuse:
		idx = 3
	default:
		fmt.Printf("err: incompatible mode and profile type. %s & %s\n", ReadProfileType(p), mode)
		os.Exit(-1)
	}

	for _, pprofSample := range p.Sample {
		var sample Sample
		// cpu [1] cpu usage
		// space [1] heap allocated
		// space [3] heap in use
		value := pprofSample.GetValue()[idx]

		for i := len(pprofSample.LocationId) - 1; i >= 0; i-- {
			l := pprofSample.LocationId[i]
			sample.Functions = append(sample.Functions, functionsMapByLocation[l]...)
			sample.Value = value
		}

		// compute the self time for the leaf
		leaf := sample.Functions[len(sample.Functions)-1]
		leaf.Self += value
		sample.Functions[len(sample.Functions)-1] = leaf

		samples = append(samples, sample)

	}

	// compute the total sampling time
	var totalSum uint64
	for _, s := range samples {
		totalSum += uint64(s.Value)
	}

	// compute the percentage for every sample
	for i, s := range samples {
		s.PercentTotal = float64(s.Value) / (float64(totalSum)) * 100.0
		samples[i] = s
	}

	return &Profile{
		Samples:         samples,
		TotalSampling:   totalSum,
		CaptureDuration: time.Duration(p.GetDurationNanos()),

		functionsMapByLocation: functionsMapByLocation,
		locationsMap:           locationsMap,
		stringsMap:             stringsMap,
	}
}

func (p *Profile) BuildTree(treeName string, aggregateByFunction bool, searchField string) *FunctionsTree {
	// prepare the tree
	tree := NewFunctionsTree(treeName)

	// fill the tree
	for _, s := range p.Samples {
		node := tree.root
		for _, f := range s.Functions {
			if s.Value == 0 {
				continue
			}
			node = node.AddFunction(f, s.Value, f.Self, s.PercentTotal, aggregateByFunction)
		}
	}

	if tree.root != nil {
		tree.root.filter(searchField)
	}

	tree.sort()

	return tree
}

func buildLocationsMap(profile *pprof.Profile, stringsMap StringsMap, functionsMap FunctionsMap) (LocationsMap, ManyFunctionsMap) {
	rv := make(LocationsMap)
	lrv := make(ManyFunctionsMap)

	for _, location := range profile.Location {
		if location.Line[0] == nil {
			continue
		}

		loc := Location{}

		for idx := len(location.Line) - 1; idx >= 0; idx-- {
			line := location.Line[idx]
			inlined := idx != len(location.Line)-1

			f := functionsMap[line.GetFunctionId()]
			f.LineNumber = uint64(line.GetLine())
			if inlined {
				f.Name = fmt.Sprintf("(inlined) %s", f.Name)
			}
			loc.Functions = append(loc.Functions, f)

			// set the line number in functions map if not inlined
			if !inlined {
				f := functionsMap[line.GetFunctionId()]
				f.LineNumber = uint64(line.GetLine())
				functionsMap[line.GetFunctionId()] = f
			}

			fs := lrv[uint64(location.GetId())]
			lrv[uint64(location.GetId())] = append(fs, f)

			rv[uint64(location.GetId())] = loc
		}
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
