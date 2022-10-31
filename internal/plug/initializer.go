package plug

type Initializer interface {
	Init(cc InitContext) error
}
