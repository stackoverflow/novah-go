package ast

import (
	"fmt"

	"github.com/stackoverflow/novah-go/data"
)

type LabelMap[T any] []data.Tuple[string, T]

func ShowLabelMap[T fmt.Stringer](lm LabelMap[T]) string {
	return fmt.Sprintf("{%s}", ShowLabels(lm, func(k string, v T) string {
		return fmt.Sprintf("%s: %s", k, v.String())
	}))
}

func showLabel(label string) string {
	return label
}

func ShowLabels[V any](labels LabelMap[V], f func(string, V) string) string {
	return data.JoinToString(labels, ", ", func(tu data.Tuple[string, V]) string {
		return f(showLabel(tu.V1), tu.V2)
	})
}
