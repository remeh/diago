package main

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
	"github.com/dustin/go-humanize"

	"github.com/remeh/diago/pprof"
)

type GUI struct {
	// data
	pprofProfile *pprof.Profile
	profile      *Profile
	tree         *FunctionsTree

	// ui options
	mode                guiMode
	profileType         string
	searchField         string
	aggregateByFunction bool
}

type guiMode string

var (
	ModeCpu       guiMode = "cpu"
	ModeHeapAlloc guiMode = "heap-alloc"
	ModeHeapInuse guiMode = "heap-inuse"
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

	if g.profileType == "cpu" {
		g.mode = ModeCpu
	} else {
		g.mode = ModeHeapAlloc
	}

	return g
}

func (g *GUI) OpenWindow() {
	wnd := giu.NewMasterWindow("Diago", 800, 600, 0, nil)
	wnd.Main(g.windowLoop)
}

func (g *GUI) onAggregationClick() {
	g.reloadProfile()
}

func (g *GUI) onSearch() {
	g.tree = g.profile.BuildTree(config.File, g.aggregateByFunction, g.searchField)
}

func (g *GUI) reloadProfile() {
	// read the pprof profile
	// ----------------------

	profile, err := NewProfile(g.pprofProfile)
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
	size := giu.Context.GetPlatform().DisplaySize()
	scale := giu.Context.GetPlatform().GetContentScale()
	giu.SingleWindow("Diago", giu.Layout{
		giu.Line(
			giu.InputTextV("filter...", size[0]/4/scale, &g.searchField, imgui.InputTextFlagsCallbackAlways, nil, g.onSearch),
			giu.Checkbox("aggregate by functions", &g.aggregateByFunction, g.onAggregationClick),
			giu.Tooltip("By default, Diago aggregates by functions, uncheck to have the information up to the lines of code"),
		),
		g.treeFromFunctionsTree(g.tree),
	})
}

func (g *GUI) treeFromFunctionsTree(tree *FunctionsTree) giu.Layout {
	// generate the header
	// ----------------------

	var text string
	if g.profile.Type == "cpu" {
		text = fmt.Sprintf("%s - total sampling duration: %s - total capture duration %s", tree.name, time.Duration(g.profile.TotalSampling).String(), g.profile.CaptureDuration.String())
	} else {
		text = fmt.Sprintf("%s - total allocated memory: %s", tree.name, humanize.IBytes(g.profile.TotalSampling))
	}

	// start generating the tree
	// ----------------------

	return giu.Layout{
		giu.Line(
			giu.TreeNode(
				text,
				giu.TreeNodeFlagsNone|giu.TreeNodeFlagsFramed|giu.TreeNodeFlagsDefaultOpen,
				g.treeNodeFromFunctionsTreeNode(tree.root),
			),
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
		_, tooltip, lineText := g.texts(child)

		// append the line to the tree
		// ----------------------

		scale := giu.Context.GetPlatform().GetContentScale()
		rv = append(rv, giu.Line(
			giu.ProgressBar(float32(child.percent)/100, 90/scale, 0, fmt.Sprintf("%.3f%%", child.percent)),
			giu.Tooltip(tooltip),
			giu.TreeNode(
				lineText,
				flags,
				g.treeNodeFromFunctionsTreeNode(child),
			),
		))
	}

	return rv
}

func (g *GUI) texts(node *treeNode) (value string, tooltip string, lineText string) {
	if g.profile.Type == "cpu" {
		value = time.Duration(node.value).String()
		tooltip = fmt.Sprintf("%s of %s", value, time.Duration(g.profile.TotalSampling).String())
	} else {
		value = humanize.IBytes(uint64(node.value))
		tooltip = fmt.Sprintf("%s of %s", value, humanize.IBytes(g.profile.TotalSampling))
	}
	lineText = fmt.Sprintf("%s %s:%d - %s", node.function.Name, path.Base(node.function.File), node.function.LineNumber, value)
	if g.aggregateByFunction {
		lineText = fmt.Sprintf("%s %s - %s", node.function.Name, path.Base(node.function.File), value)
	}
	return value, tooltip, lineText
}
