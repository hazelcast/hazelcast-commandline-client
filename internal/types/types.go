package types

type Tuple2[A, B any] struct {
	First  A
	Second B
}

func MakeTuple2[A, B any](first A, second B) Tuple2[A, B] {
	return Tuple2[A, B]{
		First:  first,
		Second: second,
	}
}

type Set[K comparable] struct {
	m map[K]struct{}
}

// MakeSet creates a set with the given items.
func MakeSet[K comparable](items ...K) Set[K] {
	s := Set[K]{
		m: map[K]struct{}{},
	}
	for _, v := range items {
		s.Add(v)
	}
	return s
}

// Add adds the given item to the set.
func (s *Set[K]) Add(item K) {
	s.m[item] = struct{}{}
}

func (s *Set[K]) Delete(item K) {
	delete(s.m, item)
}

func (s *Set[K]) Map() map[K]struct{} {
	return s.m
}

func (s *Set[K]) Has(item K) bool {
	_, ok := s.m[item]
	return ok
}

// Len returns the number of items in the set.
func (s *Set[K]) Len() int {
	return len(s.m)
}

// Diff returns a set with items which are not in other.
func (s Set[K]) Diff(other Set[K]) Set[K] {
	r := MakeSet[K]()
	for my := range s.m {
		if !other.Has(my) {
			r.Add(my)
		}
	}
	return r
}

// Items returns the items in the set.
// Returned items are not sorted.
func (s Set[K]) Items() []K {
	r := make([]K, 0, len(s.m))
	for item := range s.m {
		r = append(r, item)
	}
	return r
}
