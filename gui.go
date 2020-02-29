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
	pprofProfile *pprof.Profile
	profile      *Profile
	tree         *FunctionsTree

	profileType         string
	searchField         string
	aggregateByFunction bool
}

func NewGUI(profile *pprof.Profile) *GUI {
	g := &GUI{
		pprofProfile:        profile,
		aggregateByFunction: true,
	}
	g.reloadProfile()
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
	profile, err := NewProfile(g.pprofProfile)
	if err != nil {
		fmt.Println("err:", err)
		os.Exit(-1)
	}
	g.profile = profile
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
	var text string
	if g.profile.Type == "cpu" {
		text = fmt.Sprintf("%s - total sampling duration: %s - total capture duration %s", tree.name, time.Duration(g.profile.TotalSampling).String(), g.profile.CaptureDuration.String())
	} else {
		text = fmt.Sprintf("%s - total allocated memory: %s", tree.name, humanize.IBytes(g.profile.TotalSampling))
	}

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

		var value, tooltip string
		if g.profile.Type == "cpu" {
			value = time.Duration(child.value).String()
			tooltip = fmt.Sprintf("%s of %s", value, time.Duration(g.profile.TotalSampling).String())
		} else {
			value = humanize.IBytes(uint64(child.value))
			tooltip = fmt.Sprintf("%s of %s", value, humanize.IBytes(g.profile.TotalSampling))
		}

		lineText := fmt.Sprintf("%s %s:%d - %s", child.function.Name, path.Base(child.function.File), child.function.LineNumber, value)
		if g.aggregateByFunction {
			lineText = fmt.Sprintf("%s %s - %s", child.function.Name, path.Base(child.function.File), value)
		}

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
