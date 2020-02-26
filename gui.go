package main

import (
	"fmt"
	"path"

	"github.com/AllenDang/giu"
	"github.com/AllenDang/giu/imgui"
)

type GUI struct {
	profile *Profile
	tree    *FunctionsTree

	searchField         string
	aggregateByFunction bool
}

func NewGUI(profile *Profile, tree *FunctionsTree) *GUI {
	return &GUI{
		profile:             profile,
		tree:                tree,
		aggregateByFunction: true,
	}
}

func (g *GUI) OpenWindow() {
	wnd := giu.NewMasterWindow("Diago", 800, 600, 0, nil)
	wnd.Main(g.windowLoop)
}

func (g *GUI) onAggregationClick() {
	g.buildTree()
}

func (g *GUI) onSearch() {
	g.buildTree()
}

func (g *GUI) buildTree() {
	fmt.Println("Re-building the tree. Aggregation by functions:", g.aggregateByFunction)
	g.tree = g.profile.BuildTree(config.File, g.aggregateByFunction, g.searchField)
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
	return giu.Layout{
		giu.Line(
			giu.TreeNode(
				fmt.Sprintf("%s - total sampling duration: %s - total capture duration %s", tree.name, g.profile.SamplingDuration.String(), g.profile.CaptureDuration.String()),
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

		lineText := fmt.Sprintf("%s %s:%d - %s", child.function.Name, path.Base(child.function.File), child.function.LineNumber, child.value.String())
		if g.aggregateByFunction {
			lineText = fmt.Sprintf("%s %s - %s", child.function.Name, path.Base(child.function.File), child.value.String())
		}

		scale := giu.Context.GetPlatform().GetContentScale()
		rv = append(rv, giu.Line(
			giu.ProgressBar(float32(child.percent)/100, 90/scale, 0, fmt.Sprintf("%.3f%%", child.percent)),
			giu.Tooltip(fmt.Sprintf("%s of %s", child.value.String(), g.profile.SamplingDuration.String())),
			giu.TreeNode(
				lineText,
				flags,
				g.treeNodeFromFunctionsTreeNode(child),
			),
		))
	}

	return rv
}
