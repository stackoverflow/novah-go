package data

import (
	"fmt"
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

func (o *Option[T]) IsEmpty() bool {
	return o.empty
}

func (s *Option[T]) Value() T {
	return s.val
}

func (s *Option[T]) ValueOrNil() *T {
	if s.empty {
		return nil
	}
	return &s.val
}

///////////////////////////////////////////
// Result
///////////////////////////////////////////

type Result[T any] struct {
	Okay  T
	Error error
	isErr bool
}

func Ok[T any](v T) Result[T] {
	return Result[T]{Okay: v, isErr: false}
}

func Err[T any](err error) Result[T] {
	return Result[T]{Error: err, isErr: true}
}

func (r *Result[T]) IsOk() bool {
	return r.isErr == false
}

func (r *Result[T]) IsErr() bool {
	return r.isErr == true
}

// others

func JoinToStringStr(l []string, sep string) string {
	return JoinToStringFunc(l, sep, func(x string) string { return x })
}

func JoinToString[T fmt.Stringer](l []T, sep string) string {
	return JoinToStringFunc(l, sep, func(x T) string { return x.String() })
}

func JoinToStringFunc[T any](l []T, sep string, fun func(T) string) string {
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

// Prepends the given ident before each line of the string
func PrependIdent(str, ident string) string {
	strs := strings.FieldsFunc(str, func(r rune) bool { return r == '\n' })
	var bd strings.Builder
	for i, s := range strs {
		if i != 0 {
			bd.WriteRune('\n')
		}
		bd.WriteString(ident)
		bd.WriteString(s)
	}
	return bd.String()
}
