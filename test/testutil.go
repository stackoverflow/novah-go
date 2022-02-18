package test

import "testing"

func Equals[T comparable](t *testing.T, got, exp T) {
	if exp != got {
		t.Errorf("Result was incorrect: expected %v, got %v", exp, got)
	}
}
