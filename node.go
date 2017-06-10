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

	find(path string, m foundMatcher) (found node, vars []*Variable)

	put(handler http.Handler, methods ...string)
	get(cleanedMethod string) (http.Handler, error)
	isRegistered() bool
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
	return newInsertStatic(&n.staticChildren[index], n.staticChildren[index], static)
}

func newInsertStatic(parent **staticNode, toSplit *staticNode, static string) (node, error) {
	prefix := muxpath.CommonPrefix(toSplit.value, static)
	if len(prefix) == 0 {
		return nil, errInvalidState
	}
	if len(prefix) == len(toSplit.value) {
		return toSplit.appendStatic(static[len(prefix):])
	}

	newNode := &staticNode{
		value:          prefix,
		staticChildren: []*staticNode{toSplit},
	}
	*parent = newNode
	toSplit.value = toSplit.value[len(prefix):]

	remaining := static[len(prefix):]
	if len(remaining) == 0 {
		return newNode, nil
	}
	return newNode.appendStatic(remaining)
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

func (n *staticNode) find(path string, m foundMatcher) (node, []*Variable) {
	if prefix := muxpath.CommonPrefix(path, n.value); len(prefix) != len(n.value) {
		return nil, nil
	}

	remaining := path[len(n.value):]

	//if true then child must be segment variable
	if n.segmentVarChild != nil {
		return n.maybeFindSegmentVarChild(remaining, m)
	}
	//now we know either static or end variable child

	found, vars := n.findStaticChildDescendant(remaining, m)
	if found != nil {
		return found, vars
	}

	if m.matches(n, remaining) {
		return n, nil
	}

	if n.endVarChild != nil {
		return n.endVarChild.find(remaining, m)
	}

	return nil, nil
}

func (n *staticNode) findStaticChildDescendant(path string, m foundMatcher) (node, []*Variable) {
	index := n.indexOfCommonPrefixChild(path)

	if index < 0 {
		return nil, nil
	}

	return n.staticChildren[index].find(path, m)
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

//n.segmentVarChild must not be nil.
func (n *staticNode) maybeFindSegmentVarChild(path string, m foundMatcher) (node, []*Variable) {
	if len(path) == 0 && strings.HasSuffix(n.value, muxpath.Slash) {
		return nil, nil
	}
	return n.segmentVarChild.find(path, m)
}

//n.endVarChild must not be nil.
func (n *staticNode) maybeFindEndVarChild(path string, m foundMatcher) (node, []*Variable) {
	if len(path) == 0 {
		return nil, nil
	}
	return n.endVarChild.find(path, m)
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

	// return n.staticChild.insertStatic(n.staticChild, static)
	return newInsertStatic(&n.staticChild, n.staticChild, static)
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

func (n *segmentVarNode) find(path string, m foundMatcher) (node, []*Variable) {
	index := strings.IndexRune(path, muxpath.SlashRune)
	if index < 0 {
		index = len(path)
	}

	vars := []*Variable{
		&Variable{
			Name:  VarName(n.name),
			Value: path[:index],
		},
	}

	remaining := path[index:]

	if n.staticChild != nil {
		var childVars []*Variable = nil
		found, childVars := n.staticChild.find(remaining, m)

		if found != nil {
			return found, append(vars, childVars...)
		}
	}

	if m.matches(n, remaining) {
		return n, vars
	}

	return nil, nil
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

func (n *endVarNode) find(path string, _ foundMatcher) (node, []*Variable) {
	return n, []*Variable{
		&Variable{
			Name:  VarName(n.name),
			Value: path,
		},
	}
}

type foundMatcher interface {
	matches(n node, remaining string) bool
}

type stringFoundMatcher string

func (m stringFoundMatcher) matches(n node, remaining string) bool {
	return n != nil && n.isRegistered() && (len(remaining) == 0 || remaining == string(m))
}
