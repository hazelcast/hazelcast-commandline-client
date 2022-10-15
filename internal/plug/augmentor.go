package plug

type Augmentor interface {
	Augment(ec ExecContext, props *Properties) error
}
