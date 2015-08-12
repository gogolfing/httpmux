package path

import (
	pathlib "path"
	"regexp"
)

type PathType uint8

const (
	PathTypeStatic PathType = iota
	PathTypePartVariable
	PathTypeEndVariable
)

func TypeOf(path string) PathType {
	if IsEndVariable(path) {
		return PathTypeEndVariable
	}
	if IsVariable(path) {
		return PathTypePartVariable
	}
	return PathTypeStatic
}

func (pt PathType) IsVariable() bool {
	return pt.IsPartVariable() || pt.IsEndVariable()
}

func (pt PathType) IsPartVariable() bool {
	return pt == PathTypePartVariable
}

func (pt PathType) IsEndVariable() bool {
	return pt == PathTypeEndVariable
}

var varRegexp *regexp.Regexp

func init() {
	varRegexp = regexp.MustCompile(`\{\*?[A-Za-z]+\}`)
}

func SplitPathVars(path string) []string {
	indices := FindAllVarSubmatchIndex(path)
	if len(indices) == 0 {
		return []string{path}
	}
	result := []string{}
	last := []int{-1, 0}
	for _, v := range indices {
		result = append(result, path[last[1]:v[0]])
		result = append(result, path[v[0]:v[1]])
		last = v
	}
	result = append(result, path[last[1]:])
	if len(result[0]) == 0 {
		result = result[1:]
	}
	if len(result[len(result)-1]) == 0 {
		result = result[:len(result)-1]
	}
	return result
}

func IsEndVariable(path string) bool {
	return IsVariable(path) && path[1] == '*'
}

func IsVariable(path string) bool {
	indices := FindAllVarSubmatchIndex(path)
	return len(indices) == 1 && indices[0][0] == 0 && indices[0][1] == len(path)
}

func FindAllVarSubmatchIndex(path string) [][]int {
	indices := varRegexp.FindAllStringSubmatchIndex(path, -1)
	return indices
}

func Clean(path string) string {
	path = EnsureRootSlash(path)
	newPath := pathlib.Clean(path)
	if path[len(path)-1] == '/' && newPath != "/" {
		newPath += "/"
	}
	return newPath
}

func EnsureRootSlash(path string) string {
	if len(path) == 0 {
		return "/"
	}
	if path[0] != '/' {
		return "/" + path
	}
	return path
}

func CommonPrefix(a, b string) string {
	result := []byte{}
	for i := 0; i < len(a) && i < len(b); i++ {
		if a[i] == b[i] {
			result = append(result, a[i])
		} else {
			break
		}
	}
	return string(result)
}

func CompareIgnorePrefix(a, b string) (int, string) {
	if len(a) == 0 || len(b) == 0 {
		return len(a) - len(b), ""
	}
	i := 0
	for ; i < len(a) && i < len(b) && a[i] == b[i]; i++ {
	}
	if i == len(a) {
		return 0, a
	}
	if i == len(b) {
		return 0, b
	}
	return int(a[i]) - int(b[i]), a[:i]
}
