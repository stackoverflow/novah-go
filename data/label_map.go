package data

import (
	"fmt"
	"regexp"
)

type Entry[T any] struct {
	Label string
	Val   T
}

type LabelMap[T any] struct {
	labels []Entry[T]
}

func EmptyLabelMap[T any]() LabelMap[T] {
	return LabelMap[T]{labels: []Entry[T]{}}
}

func LabelMapFrom[T any](labels []Entry[T]) LabelMap[T] {
	return LabelMap[T]{labels}
}

func ShowRaw[T fmt.Stringer](lm LabelMap[T]) string {
	return fmt.Sprintf("{%s}", lm.Show(func(k string, v T) string {
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

func (lm LabelMap[T]) Show(f func(string, T) string) string {
	return JoinToStringFunc(lm.labels, ", ", func(tu Entry[T]) string {
		return f(ShowLabel(tu.Label), tu.Val)
	})
}

func (lm LabelMap[T]) Values() []T {
	res := make([]T, len(lm.labels))
	for i, tu := range lm.labels {
		res[i] = tu.Val
	}
	return res
}

func LabelMapValues[T, R any](lm LabelMap[T], mapper func(T) R) LabelMap[R] {
	res := make([]Entry[R], len(lm.labels))
	for i, elem := range lm.labels {
		res[i] = Entry[R]{Label: elem.Label, Val: mapper(elem.Val)}
	}
	return LabelMap[R]{labels: res}
}

func LabelFlatMapValues[T, R any](lm LabelMap[T], mapper func(T) []R) []R {
	res := make([]R, 0, len(lm.labels))
	for _, elem := range lm.labels {
		res = append(res, mapper(elem.Val)...)
	}
	return res
}

// Like LabelMapValues but short circuits on error
func LabelMapValuesErr[T, R any](lm LabelMap[T], mapper func(T) (R, error)) (LabelMap[R], error) {
	res := make([]Entry[R], len(lm.labels))
	for i, elem := range lm.labels {
		val, err := mapper(elem.Val)
		if err != nil {
			return LabelMap[R]{}, err
		}
		res[i] = Entry[R]{Label: elem.Label, Val: val}
	}
	return LabelMap[R]{labels: res}, nil
}
