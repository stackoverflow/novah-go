package data

import "testing"

func TestSet(t *testing.T) {
	s := NewSet(5, 6, 7, 8)

	if !s.Contains(5) {
		t.Error("Set should contain 5")
	}
	if !s.Contains(8) {
		t.Error("Set should contain 8")
	}

	s.Remove(5)
	if s.Contains(5) {
		t.Error("Set should not contain 5")
	}
	s.Add(10)
	if !s.Contains(10) {
		t.Error("Set should contain 10")
	}

	s2 := s.Copy()
	s2.Add(12)
	s2.Remove(8)
	if s.Contains(12) {
		t.Error("Clone failed")
	}
	if !s.Contains(8) {
		t.Error("Clone failed")
	}
}
