package internal

// Iterator is a generic iterator interface.
// Non thread safe.
type Iterator[T any] interface {
	// Next returns false if the iterator is exhausted.
	// Otherwise advances the iterator and returns true.
	Next() bool
	// Value returns the current value in the iterator.
	// Next should always be called before Value is called.
	// Otherwise may panic.
	Value() T
	// Err contains the error after advancing the iterator.
	// If it is nil, it is safe to call Next.
	// Otherwise Next should not be called.
	Err() error
}
