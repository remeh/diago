package main

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/AllenDang/giu"
	"github.com/AllenDang/imgui-go"
	"github.com/dustin/go-humanize"

	"github.com/remeh/diago/pprof"
)

type GUI struct {
	// data
	pprofProfile *pprof.Profile
	profile      *Profile
	tree         *FunctionsTree

	// ui options
	mode                sampleMode
	searchField         string
	aggregateByFunction bool
}

type sampleMode string

var (
	// use this when you don't really know the mode
	// to use to read the profile.
	ModeDefault   sampleMode = ""
	ModeCpu       sampleMode = "cpu"
	ModeHeapAlloc sampleMode = "heap-alloc"
	ModeHeapInuse sampleMode = "heap-inuse"
)

func NewGUI(profile *pprof.Profile) *GUI {
	// init the base GUI object and load the profile
	// ----------------------

	g := &GUI{
		pprofProfile:        profile,
		aggregateByFunction: true,
	}
	g.reloadProfile()

	// depending on the profile opened, switch the either the ModeCpu
	// or the ModeHeapAlloc.
	// ----------------------

	switch ReadProfileType(profile) {
	case "space":
		g.mode = ModeHeapAlloc
	case "cpu":
		g.mode = ModeCpu
	default:
		g.mode = ModeDefault
	}

	return g
}

func (g *GUI) OpenWindow() {
	wnd := giu.NewMasterWindow("Diago", 800, 600, 0)
	wnd.Run(g.windowLoop)
}

func (g *GUI) onAggregationClick() {
	g.reloadProfile()
}

func (g *GUI) onAllocated() {
	g.mode = ModeHeapAlloc
	g.reloadProfile()
}

func (g *GUI) onInuse() {
	g.mode = ModeHeapInuse
	g.reloadProfile()
}

func (g *GUI) onSearch() {
	g.tree = g.profile.BuildTree(config.File, g.aggregateByFunction, g.searchField)
}

func (g *GUI) reloadProfile() {
	// read the pprof profile
	// ----------------------

	profile, err := NewProfile(g.pprofProfile, g.mode)
	if err != nil {
		fmt.Println("err:", err)
		os.Exit(-1)
	}
	g.profile = profile

	// rebuild the displayed tree
	// ----------------------

	g.tree = profile.BuildTree(config.File, g.aggregateByFunction, g.searchField)
}

func (g *GUI) windowLoop() {
	giu.SingleWindow().Layout(
		g.toolbox(),
		g.treeFromFunctionsTree(g.tree),
	)
}

func (g *GUI) toolbox() *giu.RowWidget {
	size := giu.Context.GetPlatform().DisplaySize()

	widgets := make([]giu.Widget, 0)

	// search bar
	// ----------------------

	filterText := giu.InputText(&g.searchField).Flags(imgui.InputTextFlagsCallbackAlways).Label("Filter...").OnChange(g.onSearch).Size(size[0] / 4)

	widgets = append(widgets, filterText)

	// aggregate per func option
	// ----------------------
	widgets = append(widgets,
		giu.Checkbox("aggregate by functions", &g.aggregateByFunction).OnChange(g.onAggregationClick))
	widgets = append(widgets,
		giu.Tooltip("By default, Diago aggregates by functions, uncheck to have the information up to the lines of code"))

	// in heap mode, offer the two modes
	// ----------------------
	if g.mode == ModeHeapAlloc || g.mode == ModeHeapInuse {
		widgets = append(widgets,
			giu.RadioButton("allocated", g.mode == ModeHeapAlloc).OnChange(g.onAllocated))
		widgets = append(widgets,
			giu.RadioButton("inuse", g.mode == ModeHeapInuse).OnChange(g.onInuse))
	}

	return giu.Row(
		widgets...,
	)
}

func (g *GUI) treeFromFunctionsTree(tree *FunctionsTree) giu.Layout {
	// generate the header
	// ----------------------

	var text string
	switch g.mode {
	case ModeCpu:
		text = fmt.Sprintf("%s - total sampling duration: %s - total capture duration %s", tree.name, time.Duration(g.profile.TotalSampling).String(), g.profile.CaptureDuration.String())
	case ModeHeapAlloc:
		text = fmt.Sprintf("%s - total allocated memory: %s", tree.name, humanize.IBytes(g.profile.TotalSampling))
	case ModeHeapInuse:
		text = fmt.Sprintf("%s - total in-use memory: %s", tree.name, humanize.IBytes(g.profile.TotalSampling))
	}

	// start generating the tree
	// ----------------------

	return giu.Layout{
		giu.Row(
			giu.TreeNode(text).Flags(giu.TreeNodeFlagsNone | giu.TreeNodeFlagsFramed | giu.TreeNodeFlagsDefaultOpen).Layout(g.treeNodeFromFunctionsTreeNode(tree.root)),
		),
	}
}

func (g *GUI) treeNodeFromFunctionsTreeNode(node *treeNode) giu.Layout {
	if node == nil {
		return nil
	}
	rv := giu.Layout{}
	for _, child := range node.children {
		if !child.visible {
			continue
		}

		flags := giu.TreeNodeFlagsSpanAvailWidth
		if child.isLeaf() {
			flags |= giu.TreeNodeFlagsLeaf
		}

		// generate the displayed texts
		// ----------------------
		_, _, tooltip, lineText := g.texts(child)

		// append the line to the tree
		// ----------------------

		rv = append(rv, giu.Row(
			giu.ProgressBar(float32(child.percent)/100).Size(90, 0).Overlayf("%.3f%%", child.percent),
			giu.Tooltip(tooltip),
			giu.TreeNode(lineText).Flags(flags).Layout(g.treeNodeFromFunctionsTreeNode(child)),
		),
		)
	}

	return rv
}

func (g *GUI) texts(node *treeNode) (value string, self string, tooltip string, lineText string) {
	if g.profile.Type == "cpu" {
		value = time.Duration(node.value).String()
		self = time.Duration(node.self).String()
		tooltip = fmt.Sprintf("%s of %s\nself: %s", value, time.Duration(g.profile.TotalSampling).String(), self)
	} else {
		value = humanize.IBytes(uint64(node.value))
		self = humanize.IBytes(uint64(node.self))
		tooltip = fmt.Sprintf("%s of %s\nself: %s", value, humanize.IBytes(g.profile.TotalSampling), self)
	}
	lineText = fmt.Sprintf("%s %s:%d - %s - self: %s", node.function.Name, path.Base(node.function.File), node.function.LineNumber, value, self)
	if g.aggregateByFunction {
		lineText = fmt.Sprintf("%s %s - %s - self: %s", node.function.Name, path.Base(node.function.File), value, self)
	}
	return value, self, tooltip, lineText
}
