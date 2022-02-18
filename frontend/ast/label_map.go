package ast

import (
	"fmt"
	"regexp"

	"github.com/stackoverflow/novah-go/data"
)

type LabelMap[T any] []data.Tuple[string, T]

func EmptyLabelMap[T any]() LabelMap[T] {
	return []data.Tuple[string, T]{}
}

func ShowLabelMap[T fmt.Stringer](lm LabelMap[T]) string {
	return fmt.Sprintf("{%s}", ShowLabels(lm, func(k string, v T) string {
		return fmt.Sprintf("%s: %s", k, v.String())
	}))
}

var labelRegex = regexp.MustCompile("^[a-z](?:\\w+|_)*$")

func showLabel(label string) string {
	if labelRegex.MatchString(label) {
		return label
	}
	return fmt.Sprintf(`"%s"`, label)
}

func ShowLabels[V any](labels LabelMap[V], f func(string, V) string) string {
	return data.JoinToStringFunc(labels, ", ", func(tu data.Tuple[string, V]) string {
		return f(showLabel(tu.V1), tu.V2)
	})
}

func LabelValues[T any](m *LabelMap[T]) []T {
	res := make([]T, len(*m))
	for _, tu := range *m {
		res = append(res, tu.V2)
	}
	return res
}
