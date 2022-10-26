package output

type OrderedSet[T comparable] struct {
	orderedItems   []T
	itemSet        map[T]int
	deletedIndices map[int]struct{}
}

func NewOrderedSet[T comparable]() *OrderedSet[T] {
	return &OrderedSet[T]{
		itemSet:        map[T]int{},
		deletedIndices: map[int]struct{}{},
	}
}

// Add adds an item to the ordered set.
// nil values are not accepted
func (s *OrderedSet[T]) Add(item T) bool {
	if _, ok := s.itemSet[item]; ok {
		return false
	}
	s.itemSet[item] = len(s.orderedItems)
	s.orderedItems = append(s.orderedItems, item)
	return true
}

func (s *OrderedSet[T]) Delete(item T) bool {
	if index, ok := s.itemSet[item]; ok {
		delete(s.itemSet, item)
		s.deletedIndices[index] = struct{}{}
		return true
	}
	return false
}

func (s *OrderedSet[T]) Contains(item T) bool {
	_, ok := s.itemSet[item]
	return ok
}

func (s *OrderedSet[T]) Items() []T {
	r := make([]T, 0, len(s.orderedItems))
	for index, item := range s.orderedItems {
		if _, ok := s.deletedIndices[index]; !ok {
			r = append(r, item)
		}
	}
	return r
}

func (s *OrderedSet[T]) Len() int {
	return len(s.itemSet)
}
