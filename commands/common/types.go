package common

type NamePersister interface {
	Set(name string, value string)
	Get(name string) (string, bool)
	Reset(name string)
	PersistenceInfo() map[string]string
}
