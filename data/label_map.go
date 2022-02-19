package data

import (
	"fmt"
	"regexp"
)

type LabelMap[T any] []Tuple[string, T]

func EmptyLabelMap[T any]() LabelMap[T] {
	return []Tuple[string, T]{}
}

func ShowLabelMap[T fmt.Stringer](lm LabelMap[T]) string {
	return fmt.Sprintf("{%s}", ShowLabels(lm, func(k string, v T) string {
		return fmt.Sprintf("%s: %s", k, v.String())
	}))
}

var labelRegex = regexp.MustCompile("^[a-z](?:\\w+|_)*$")

func ShowLabel(label string) string {
	if labelRegex.MatchString(label) {
		return label
	}
	return fmt.Sprintf(`"%s"`, label)
}

func ShowLabels[V any](labels LabelMap[V], f func(string, V) string) string {
	return JoinToStringFunc(labels, ", ", func(tu Tuple[string, V]) string {
		return f(ShowLabel(tu.V1), tu.V2)
	})
}

func LabelValues[T any](m *LabelMap[T]) []T {
	res := make([]T, len(*m))
	for i, tu := range *m {
		res[i] = tu.V2
	}
	return res
}

func LabelMapVals[T, R any](m LabelMap[T], mapper func(string, T) R) LabelMap[R] {
	res := make([]Tuple[string, R], len(m))
	for i, tu := range m {
		res[i] = Tuple[string, R]{V1: tu.V1, V2: mapper(tu.V1, tu.V2)}
	}
	return res
}
