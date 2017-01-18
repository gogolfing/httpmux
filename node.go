package httpmux

import (
	"fmt"
	"net/http"
	"strings"

	muxpath "github.com/gogolfing/httpmux/path"
)

var errInvalidState = fmt.Errorf("")

type node interface {
	appendStatic(static string) (node, error)
	appendSegmentVar(name VarName) (node, error)
	appendEndVar(name VarName) (node, error)

	find(path string, m foundMatcher) (found node, vars []*Variable, remaining string)

	put(handler http.Handler, methods ...string)
	get(cleanedMethod string) (http.Handler, error)
}

type staticNode struct {
	value string

	staticChildren  []*staticNode
	segmentVarChild *segmentVarNode
	endVarChild     *endVarNode

	methodHandler
}

func (n *staticNode) appendStatic(static string) (node, error) {
	if len(static) == 0 {
		return n, nil
	}
	if n.segmentVarChild != nil {
		return nil, ErrOverlapStaticVar(n.segmentVarChild.name)
	}
	index := n.indexOfCommonPrefixChild(static)
	if index < 0 { //child not found. needs to be inserted at ^index.
		newChild := &staticNode{value: static}
		n.insertStaticChildAtIndex(newChild, ^index)
		return newChild, nil
	}
	//now we know which child to insert on
	return n.staticChildren[index].insertStatic(&n, static)
}

func (n *staticNode) insertStatic(split **staticNode, static string) (node, error) {
	prefix := muxpath.CommonPrefix(static, n.value)
	if len(prefix) == 0 {
		return nil, errInvalidState
	}
	if len(prefix) == len(n.value) {
		return n.appendStatic(static[len(prefix):])
	}

	//now need to split self
	*split = &staticNode{
		value:          prefix,
		staticChildren: []*staticNode{n},
	}
	n.value = n.value[len(prefix):]

	return *split, nil
}

func (n *staticNode) insertStaticChildAtIndex(child *staticNode, index int) {
	before, after := n.staticChildren[:index], n.staticChildren[index:]
	n.staticChildren = make([]*staticNode, 0, len(before)+1+len(after))
	n.staticChildren = append(n.staticChildren, before...)
	n.staticChildren = append(n.staticChildren, child)
	n.staticChildren = append(n.staticChildren, after...)
}

func (n *staticNode) appendSegmentVar(name VarName) (node, error) {
	if len(n.staticChildren) > 0 { //static case
		return nil, ErrOverlapStaticVar(name)
	}
	if n.segmentVarChild == nil && n.endVarChild == nil { //empty case
		n.segmentVarChild = &segmentVarNode{name: name}
		return n.segmentVarChild, nil
	}
	if n.segmentVarChild != nil {
		if n.segmentVarChild.name != name { //unequal names
			return nil, &ErrUnequalVars{Variable1: n.segmentVarChild.name, Variable2: name}
		}
		return n.segmentVarChild, nil //otherwise names are equal so return the child
	}
	//now we must have an end variable. this is always an error.
	return nil, &ErrUnequalVars{n.endVarChild.name, name}
}

func (n *staticNode) appendEndVar(name VarName) (node, error) {
	if len(n.staticChildren) > 0 { //static case
		n.endVarChild = &endVarNode{name: name}
		return n.endVarChild, nil
	}
	if n.segmentVarChild == nil && n.endVarChild == nil { //empty case
		n.endVarChild = &endVarNode{name: name}
		return n.endVarChild, nil
	}
	if n.segmentVarChild != nil {
		return nil, &ErrUnequalVars{Variable1: n.segmentVarChild.name, Variable2: name}
	}
	if n.endVarChild.name != name { //unequal names
		return nil, &ErrUnequalVars{Variable1: n.endVarChild.name, Variable2: name}
	}
	//otherwise names are equal so return the child
	return n.endVarChild, nil
}

func (n *staticNode) find(path string, m foundMatcher) (node, []*Variable, string) {
	if prefix := muxpath.CommonPrefix(path, n.value); len(prefix) != len(n.value) {
		return n, nil, path
	}

	remaining := path[len(n.value):]

	if n.segmentVarChild != nil {
		return n.segmentVarChild.find(remaining, m)
	}

	if n.endVarChild == nil {
		return n.findStaticChildDescendant(remaining, m)
	}

	//we know endVarChild is not nil and we need to attempt static before end variable
	sdcNode, sdcVars, sdcRemaining := n.findStaticChildDescendant(remaining, m)

	if m.matches(sdcNode, sdcRemaining) {
		return sdcNode, sdcVars, sdcRemaining
	}
	return n.endVarChild.find(remaining, m)
}

func (n *staticNode) findStaticChildDescendant(path string, m foundMatcher) (node, []*Variable, string) {
	if len(n.staticChildren) > 0 {
		index := n.indexOfCommonPrefixChild(path)
		if index < 0 {
			return n, nil, path
		}
		return n.staticChildren[index].find(path, m)
	}
	return n, nil, path
}

func (n *staticNode) indexOfCommonPrefixChild(static string) int {
	low, high := 0, len(n.staticChildren)
	for low < high {
		mid := (low + high) >> 1
		comparison, prefix := muxpath.CompareIgnoringPrefix(static, n.staticChildren[mid].value)
		if len(prefix) > 0 {
			return mid
		} else if comparison < 0 {
			high = mid
		} else { //comparison must be > 0
			low = mid + 1
		}
	}
	return ^high
}

type segmentVarNode struct {
	name VarName

	staticChild *staticNode

	methodHandler
}

func (n *segmentVarNode) appendStatic(static string) (node, error) {
	if n.staticChild == nil {
		n.staticChild = &staticNode{
			value: static,
		}
		return n.staticChild, nil
	}

	return n.staticChild.insertStatic(&n.staticChild, static)
}

func (n *segmentVarNode) appendSegmentVar(name VarName) (node, error) {
	return nil, &ErrConsecutiveVars{
		Variable1: n.name,
		Variable2: name,
	}
}

func (n *segmentVarNode) appendEndVar(name VarName) (node, error) {
	return nil, &ErrConsecutiveVars{
		Variable1: n.name,
		Variable2: name,
	}
}

func (n *segmentVarNode) find(path string, m foundMatcher) (found node, vars []*Variable, remaining string) {
	index := strings.IndexRune(path, muxpath.RootPathRune)
	if index < 0 {
		index = len(path)
	}

	found = n
	vars = []*Variable{
		&Variable{
			Name:  VarName(n.name),
			Value: path[:index],
		}}
	remaining = path[index:]

	if n.staticChild != nil {
		var childVars []*Variable = nil
		found, childVars, remaining = n.find(remaining, m)
		vars = append(vars, childVars...)
	}

	return
}

type endVarNode struct {
	name VarName

	methodHandler
}

func (n *endVarNode) appendStatic(static string) (node, error) {
	return nil, errInvalidState
}

func (n *endVarNode) appendSegmentVar(name VarName) (node, error) {
	return nil, errInvalidState
}

func (n *endVarNode) appendEndVar(name VarName) (node, error) {
	return nil, errInvalidState
}

func (n *endVarNode) find(path string, _ foundMatcher) (found node, vars []*Variable, remaining string) {
	found = n
	vars = []*Variable{
		&Variable{
			Name:  VarName(n.name),
			Value: path,
		}}
	return
}

type foundMatcher interface {
	matches(n node, remaining string) bool
}

type stringFoundMatcher string

func (m stringFoundMatcher) matches(n node, remaining string) bool {
	return n != nil && (len(remaining) > 0 || remaining == string(m))
}
