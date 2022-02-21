package data

import "testing"

func TestStack(t *testing.T) {
	s := NewStack(5, 6, 7)
	if s.Size() != 3 {
		t.Error("Expected stack to have 3 elements")
	}
	s.Push(3)
	s.Push(9)
	v := s.Pop()

	if s.Size() != 4 {
		t.Error("Expected size to have changed")
	}
	if v != 9 {
		t.Error("Expected popped value to be 9")
	}
}
