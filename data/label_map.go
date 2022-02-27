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

func LabelMapFrom[T any](labels ...Entry[T]) LabelMap[T] {
	return LabelMap[T]{labels}
}

func LabelMapSingleton[T any](label string, val T) LabelMap[T] {
	return LabelMap[T]{labels: []Entry[T]{{Label: label, Val: val}}}
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

func (lm LabelMap[T]) IsEmpty() bool {
	return len(lm.labels) == 0
}

func (lm LabelMap[T]) Size() int {
	return len(lm.labels)
}

func (lm LabelMap[T]) Entries() []Entry[T] {
	return lm.labels
}

// Creates a shallow copy of the map
func (lm LabelMap[T]) Copy() LabelMap[T] {
	res := make([]Entry[T], len(lm.labels))
	copy(res, lm.labels)
	return LabelMap[T]{labels: res}
}

func (lm LabelMap[T]) Put(key string, vals []T) LabelMap[T] {
	labels := lm.labels
	for _, v := range vals {
		labels = append(labels, Entry[T]{Label: key, Val: v})
	}
	return LabelMap[T]{labels: labels}
}

// Merge the two maps together with duplicated keys
func (lm LabelMap[T]) Merge(other LabelMap[T]) LabelMap[T] {
	m := make([]Entry[T], 0, len(lm.labels)+len(other.labels))
	for _, e := range lm.labels {
		m = append(m, e)
	}
	for _, e := range other.labels {
		m = append(m, e)
	}
	return LabelMap[T]{labels: m}
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
