package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type FunctionsTree struct {
	name string
	root *treeNode
}

func (t *FunctionsTree) sort() {
	if t.root == nil {
		fmt.Println("warn: called sort() on an empty tree")
		return
	}

	t.root.sort()
}

type treeNode struct {
	children []*treeNode
	function *Function
	value    time.Duration
	percent  float64
	visible  bool
}

func NewFunctionsTree(treeName string) *FunctionsTree {
	return &FunctionsTree{
		name: treeName,
		root: &treeNode{},
	}
}

func (n treeNode) ID(lineNumber bool) string {
	if n.function == nil {
		return "Root"
	}
	return n.function.String(lineNumber)
}

// AddFunction adds the given function to the tree.
// AddFunction takes care of aggregating the values per functions calls or line of
// code depending on the aggregateByFunction parameter.
func (n *treeNode) AddFunction(f Function, value int64, percent float64, aggregateByFunction bool) *treeNode {
	for i, child := range n.children {
		// if existing, we add the values to the current node
		if child.ID(!aggregateByFunction) == f.String(!aggregateByFunction) {
			child.value = time.Duration(child.value.Nanoseconds() + value)
			child.percent += percent
			n.children[i] = child
			return child
		}
	}
	// doesn't exist, create it
	node := &treeNode{
		function: &f,
		value:    time.Duration(value),
		percent:  percent,
	}

	n.children = append(n.children, node)
	return node
}

func (n *treeNode) isLeaf() bool {
	return len(n.children) == 0
}

func (n *treeNode) filter(searchField string) bool {
	var visible bool

	if searchField == "" {
		visible = true
	} else if n.function != nil && strings.Contains(strings.ToLower(n.function.Name), strings.ToLower(searchField)) {
		visible = true
	} else if n.function != nil && strings.Contains(strings.ToLower(n.function.File), strings.ToLower(searchField)) {
		visible = true
	}

	for _, child := range n.children {
		if child.filter(searchField) {
			visible = true
		}
	}

	n.visible = visible
	return n.visible
}

func (n *treeNode) sort() {
	sort.Slice(
		n.children,
		func(i, j int) bool {
			return n.children[i].value > n.children[j].value
		},
	)
	for _, child := range n.children {
		child.sort()
	}
}
