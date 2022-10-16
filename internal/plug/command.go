package plug

type Commander interface {
	Exec(ec ExecContext) error
}
