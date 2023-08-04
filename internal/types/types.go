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
