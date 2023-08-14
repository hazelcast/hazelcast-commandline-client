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

func NewSet[K comparable](items ...K) *Set[K] {
	s := Set[K]{
		m: map[K]struct{}{},
	}
	for _, v := range items {
		s.Add(v)
	}
	return &s
}

func (s *Set[K]) Add(item K) {
	s.m[item] = struct{}{}
}

func (s *Set[K]) Has(item K) bool {
	_, ok := s.m[item]
	return ok
}

func (s *Set[K]) Len() int {
	return len(s.m)
}
