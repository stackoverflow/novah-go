package data

import (
	"strings"
)

type Tuple[T any, U any] struct {
	V1 T
	V2 U
}

///////////////////////////////////////////
// Option
///////////////////////////////////////////

type Option[T any] struct {
	val   T
	empty bool
}

func Some[T any](val T) Option[T] {
	return Option[T]{val: val, empty: false}
}

func None[T any]() Option[T] {
	return Option[T]{empty: true}
}

func (o Option[T]) Empty() bool {
	return o.empty
}

func (s Option[T]) Value() T {
	return s.val
}

func (s Option[T]) ValueOrNil() *T {
	if s.empty {
		return nil
	}
	return &s.val
}

///////////////////////////////////////////
// Result
///////////////////////////////////////////

type Result[T any] struct {
	V   *T
	err *error
}

func Ok[T any](v T) Result[T] {
	return Result[T]{&v, nil}
}

func Err[T any](err error) Result[T] {
	return Result[T]{nil, &err}
}

func (r Result[T]) IsOk() bool {
	return r.err == nil
}

func (r Result[T]) IsErr() bool {
	return r.err != nil
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
