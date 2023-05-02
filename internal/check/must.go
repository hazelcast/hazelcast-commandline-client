package check

func Must(err error) {
	if err != nil {
		panic(err)
	}
}

func MustValue[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func MustAnyValue[T any](v any, err error) T {
	if err != nil {
		panic(err)
	}
	return v.(T)
}

func MustOK[T any](v any, ok bool) T {
	if !ok {
		panic("not OK")
	}
	return v.(T)
}
