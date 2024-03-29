package data

func MapSlice[T, R any](sli []T, fun func(T) R) []R {
	res := make([]R, len(sli))
	for i, x := range sli {
		res[i] = fun(x)
	}
	return res
}

// Like MapSlice but short circuits in case of error
func MapSliceError[T, R any](sli []T, fun func(T) (R, error)) ([]R, error) {
	res := make([]R, len(sli))
	for i, x := range sli {
		elem, err := fun(x)
		if err != nil {
			return nil, err
		}
		res[i] = elem
	}
	return res, nil
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

// Returns the last index of elem in the slice
func SliceLastIndexOfFunc[T any](s []T, fun func(T) bool) int {
	index := -1
	for i, e := range s {
		if fun(e) {
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

// Returns both slices zipped together.
// The result will be the size of the smallest slice
func ZipSlices[T, R any](s1 []T, s2 []R) []Tuple[T, R] {
	size := len(s1)
	if len(s2) < size {
		size = len(s2)
	}
	res := make([]Tuple[T, R], size)
	for i := 0; i < size; i++ {
		res[i] = Tuple[T, R]{V1: s1[i], V2: s2[i]}
	}
	return res
}

// Returns a new slice with elements reversed
func ReverseSlice[T any](s []T) []T {
	a := make([]T, len(s))
	copy(a, s)

	for i := len(a)/2 - 1; i >= 0; i-- {
		opp := len(a) - 1 - i
		a[i], a[opp] = a[opp], a[i]
	}

	return a
}

// Returns true if the function is true for any element in the slice
func AnySlice[T any](s []T, pred func(T) bool) bool {
	for _, e := range s {
		if pred(e) {
			return true
		}
	}
	return false
}
