package common

import "context"

func PersisterFromContext(ctx context.Context) NamePersister {
	return ctx.Value("persister").(NamePersister)
}

func SetContext(ctx context.Context, persister NamePersister) context.Context {
	return context.WithValue(ctx, "persister", persister)
}

type NamePersister interface {
	Set(name string, value string)
	Get(name string) (string, bool)
	Reset(name string)
	PersistenceInfo() map[string]string
}
