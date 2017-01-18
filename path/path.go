package path

import (
	pathlib "path"
	"strings"
)

const (
	RootPath     = "/"
	RootPathRune = '/'

	SegmentVarRune = ':'
	EndVarRune     = '*'
)

func SplitIntoStaticAndVariableParts(path string) []string {
	result := []string{}
	static, remaining := "", path

	for len(remaining) > 0 {
		varIndex := strings.IndexAny(remaining, string([]rune{SegmentVarRune, EndVarRune}))

		switch {
		case varIndex < 0: //variable not found in rest of remaining. all the rest is static.
			static, remaining = static+remaining, ""

		case varIndex != len(remaining)-1 && remaining[varIndex] == remaining[varIndex+1]: //double variable rune. reduce to one.
			static, remaining = static+remaining[:varIndex+1], remaining[varIndex+2:]

		case remaining[varIndex] == SegmentVarRune: //found segment variable
			slashIndex := strings.IndexRune(remaining[varIndex:], RootPathRune)
			if slashIndex < 0 {
				slashIndex = len(remaining)
			} else {
				slashIndex += varIndex
			}
			if beforeVar := static + remaining[:varIndex]; len(beforeVar) > 0 {
				result = append(result, beforeVar)
			}
			result = append(result, remaining[varIndex:slashIndex])
			static, remaining = "", remaining[slashIndex:]

		case remaining[varIndex] == EndVarRune: //found end variable
			if beforeVar := static + remaining[:varIndex]; len(beforeVar) > 0 {
				result = append(result, beforeVar)
			}
			result = append(result, remaining[varIndex:])
			static, remaining = "", ""
		}
	}
	if len(static) > 0 {
		result = append(result, static)
	}

	return result
}

func staticThenVariableParts(path, static string, startIndex, varIndex, varEnd int) []string {
	if startIndex == varIndex {
		return []string{static + path[varIndex:varEnd]}
	}
	return []string{static + path[startIndex:varIndex], path[varIndex:varEnd]}
}

func ExtractVariableName(value string) (name string, ok bool) {
	if len(value) > 0 && (value[0] == SegmentVarRune || value[0] == EndVarRune) {
		return value[1:], true
	}
	return "", false
}

func IsSegmentVariable(value string) bool {
	return len(value) > 0 && value[0] == SegmentVarRune
}

func IsEndVariable(value string) bool {
	return len(value) > 0 && value[0] == EndVarRune
}

func Clean(path string) string {
	path = EnsureRootSlash(path)
	newPath := pathlib.Clean(path)
	if path[len(path)-1] == RootPathRune && newPath != RootPath {
		newPath += RootPath
	}
	return newPath
}

func EnsureRootSlash(path string) string {
	if len(path) == 0 {
		return RootPath
	}
	if path[0] != RootPathRune {
		return RootPath + path
	}
	return path
}

func CompareAfterPrefix(a, b string) (comp int, prefix string) {
	prefix = CommonPrefix(a, b)
	if a[len(prefix):] < b[len(prefix):] {
		comp = -1
	} else if a[len(prefix):] > b[len(prefix):] {
		comp = 1
	}
	return
}

func CommonPrefix(a, b string) string {
	i := 0
	for ; i < len(a) && i < len(b) && a[i] == b[i]; i++ {
	}
	return a[:i]
}

func CompareIgnoringPrefix(a, b string) (int, string) {
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
