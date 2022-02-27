package data

import "golang.org/x/exp/slices"

type Stack[T any] struct {
	vals []T
}

func NewStack[T any](vals ...T) *Stack[T] {
	return &Stack[T]{vals}
}

func (s *Stack[T]) Size() int {
	return len(s.vals)
}

func (s *Stack[T]) IsEmpty() bool {
	return len(s.vals) == 0
}

func (s *Stack[T]) Push(v T) T {
	s.vals = append(s.vals, v)
	return v
}

func (s *Stack[T]) Pop() T {
	l := len(s.vals)
	v := s.vals[l-1]
	s.vals = s.vals[:l-1]
	return v
}

func (s *Stack[T]) Peek() T {
	return s.vals[len(s.vals)-1]
}

func (s *Stack[T]) ToSlice() []T {
	return ReverseSlice(s.vals)
}

func EqualsStack[T comparable](this *Stack[T], other *Stack[T]) bool {
	return slices.Equal(this.vals, other.vals)
}
