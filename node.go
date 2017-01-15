package httpmux

import (
	"fmt"
	"strings"

	"github.com/gogolfing/httpmux/path"
)

var errInvalid = fmt.Errorf("")

type node interface {
	appendStatic(static string) (node, error)
	appendSegmentVar(name VarName) (node, error)
	appendEndVar(name VarName) (node, error)

	find(path string) (found node, v *Variable, remaining string)
}

type staticNode struct {
	value string

	staticChildren  []*staticNode
	segmentVarChild *segmentVarNode
	endVarChild     *endVarNode
}

func (n *staticNode) appendStatic(static string) (node, error) {
	return nil, nil
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

func (n *staticNode) find(path string) (found node, v *Variable, remaining string) {
	return
}

type segmentVarNode struct {
	name VarName

	child *staticNode
}

func (n *segmentVarNode) appendStatic(static string) (node, error) {
	if n.child == nil {
		n.child = &staticNode{
			value: static,
		}
		return n.child, nil
	}
	commonPrefix := path.CommonPrefix(n.child.value, static)
	if commonPrefix == "" {
		return nil, errInvalid
	}
	remaining := static[len(commonPrefix):]
	return n.child.appendStatic(remaining)
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

func (n *segmentVarNode) find(path string) (found node, v *Variable, remaining string) {
	index := strings.IndexRune(path, '/')
	if index < 0 {
		index = len(path)
	}
	found = n
	v = &Variable{
		Name:  VarName(n.name),
		Value: path[:index],
	}
	remaining = path[index:]
	return
}

type endVarNode struct {
	name VarName
}

func (n *endVarNode) appendStatic(static string) (node, error) {
	return nil, errInvalid
}

func (n *endVarNode) appendSegmentVar(name VarName) (node, error) {
	return nil, errInvalid
}

func (n *endVarNode) appendEndVar(name VarName) (node, error) {
	return nil, errInvalid
}

func (n *endVarNode) find(path string) (found node, v *Variable, remaining string) {
	found = n
	v = &Variable{
		Name:  VarName(n.name),
		Value: path,
	}
	return
}
