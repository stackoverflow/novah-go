package data

func MapSlice[T, R any](sli []T, fun func(T) R) []R {
	res := make([]R, len(sli))
	for i, x := range sli {
		res[i] = fun(x)
	}
	return res
}

func FlatMapSlice[T, R any](s []T, fun func(T) []R) []R {
	res := make([]R, 0, len(s))
	for _, par := range s {
		res = append(res, fun(par)...)
	}
	return res
}

func FilterSlice[T any](s []T, pred func(T) bool) []T {
	res := make([]T, 0, len(s))
	for _, x := range s {
		if pred(x) {
			res = append(res, x)
		}
	}
	return res
}

// Returns the first element that satisfies the predicate
// and a boolean indicating if it could find any
func FindSlice[T any](s []T, pred func(T) bool) (T, bool) {
	for _, x := range s {
		if pred(x) {
			return x, true
		}
	}
	var res T
	return res, false
}

// Filter all elements of this slice that are type R
// This runs in O(2n) because of the t -> interface{} conversion
func FilterSliceIsInstance[T, R any](s []T) []R {
	ts := make([]any, len(s))
	for i := range s {
		ts[i] = s[i]
	}

	res := make([]R, 0, len(s))
	for _, x := range ts {
		switch x.(type) {
		case R:
			res = append(res, x.(R))
		}
	}
	return res
}

// Returns the (first) index of elem in the slice
func SliceIndexOf[T comparable](s []T, elem T) int {
	for i, e := range s {
		if e == elem {
			return i
		}
	}
	return -1
}

// Returns the last index of elem in the slice
func SliceLastIndexOf[T comparable](s []T, elem T) int {
	index := -1
	for i, e := range s {
		if e == elem {
			index = i
		}
	}
	return index
}

// Delete the element at index from this slice
// requires the 0-value of this type
func DeleteAtSlice[T any](sli []T, at int, zero T) []T {
	copy(sli[at:], sli[at+1:])
	sli[len(sli)-1] = zero
	sli = sli[:len(sli)-1]
	return sli
}

// Inserts the element at index in this slice
// requires the 0-value of this type
func InsertAtSlice[T any](s []T, at int, elem T, zero T) []T {
	s = append(s, zero)
	copy(s[at+1:], s[at:])
	s[at] = elem
	return s
}

// Returns true if elem is inside the slice
func InSlice[T comparable](s []T, elem T) bool {
	for _, x := range s {
		if x == elem {
			return true
		}
	}
	return false
}

// Compares two slices for equality
func EqualsSlice[T comparable](s1 []T, s2 []T) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}
