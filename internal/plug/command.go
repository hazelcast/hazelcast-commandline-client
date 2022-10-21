package plug

type Commander interface {
	Exec(ec ExecContext) error
}

type InteractiveCommander interface {
	ExecInteractive(ec ExecInteractiveContext) error
}
