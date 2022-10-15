package plug

type Command interface {
	Init(ctx CommandContext) error
	Exec(ctx ExecContext) error
}
