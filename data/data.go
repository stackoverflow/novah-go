package data

import (
	"strings"
)

type Tuple[T any, U any] struct {
	V1 T
	V2 U
}

func JoinToString[T any](l []T, sep string, fun func(T) string) string {
	var build strings.Builder
	length := len(l)
	if length == 0 {
		return ""
	}
	build.WriteString(fun(l[0]))
	if length == 1 {
		return build.String()
	}

	for i := 1; i < length; i++ {
		build.WriteString(sep)
		build.WriteString(fun(l[i]))
	}
	return build.String()
}
