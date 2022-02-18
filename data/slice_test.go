package data

import "testing"

func TestSlice(t *testing.T) {
	s1 := []int{3, 4, 5, 6}
	s1res := MapSlice(s1, func(x int) int { return x * 10 })
	s1expect := []int{30, 40, 50, 60}
	if !EqualsSlice(s1res, s1expect) {
		t.Errorf("Result was incorrect: expected %v, got: %v", s1expect, s1res)
	}

	s2res := FilterSlice(s1, func(x int) bool { return x%2 == 0 })
	s2expect := []int{4, 6}
	if !EqualsSlice(s2res, s2expect) {
		t.Errorf("Result was incorrect: expected %v, got: %v", s2expect, s2res)
	}

	s3res := FlatMapSlice(s1, func(x int) []int { return []int{x, x * 2} })
	s3expect := []int{3, 6, 4, 8, 5, 10, 6, 12}
	if !EqualsSlice(s3res, s3expect) {
		t.Errorf("Result was incorrect: expected %v, got: %v", s3expect, s3res)
	}

	s4res, found := FindSlice(s1, func(x int) bool { return x == 4 })
	if !found || s4res != 4 {
		t.Errorf("Result was incorrect: expected 4, got: %d", s4res)
	}

	_, found2 := FindSlice(s1, func(x int) bool { return x == 10 })
	if found2 {
		t.Errorf("Result was incorrect: expected to not find element but found")
	}
}
