package plug

import "context"

type Commander interface {
	Exec(ctx context.Context, ec ExecContext) error
}

type InteractiveCommander interface {
	ExecInteractive(ctx context.Context, ec ExecContext) error
}
