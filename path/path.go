package path

import pathlib "path"

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
