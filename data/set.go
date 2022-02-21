package data

type Set[T comparable] struct {
	m map[T]bool
}

func NewSet[T comparable](elems ...T) Set[T] {
	s := Set[T]{m: make(map[T]bool)}
	s.Add(elems...)
	return s
}

func (s Set[T]) Contains(elem T) bool {
	return s.m[elem]
}

func (s Set[T]) Add(elems ...T) {
	for _, elem := range elems {
		s.m[elem] = true
	}
}

func (s Set[T]) Remove(elems ...T) {
	for _, elem := range elems {
		delete(s.m, elem)
	}
}

func (s Set[T]) Length() int {
	return len(s.m)
}

func (s Set[T]) Inner() map[T]bool {
	return s.m
}

func (s Set[T]) Equals(other Set[T]) bool {
	if len(s.m) != len(other.m) {
		return false
	}
	for k := range s.m {
		if !other.Contains(k) {
			return false
		}
	}
	return true
}

// Returns a shallow clone of the set
func (s Set[T]) Copy() Set[T] {
	newm := make(map[T]bool)
	for k, v := range s.m {
		newm[k] = v
	}
	return Set[T]{m: newm}
}
