package plug

import (
	"fmt"
	"regexp"
	"sort"

	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/mk"
)

var Registry = newRegistry()

type registry struct {
	initters        map[string]Initializer
	commands        map[string]Commander
	augmentors      map[string]Augmentor
	printers        map[string]Printer
	onlyInteractive []string
}

func newRegistry() *registry {
	return &registry{
		initters:   map[string]Initializer{},
		commands:   map[string]Commander{},
		augmentors: map[string]Augmentor{},
		printers:   map[string]Printer{},
	}
}

type RegisterCommandOption interface {
	Apply(*registry, string)
}

type OnlyInteractive struct{}

func (OnlyInteractive) Apply(r *registry, name string) {
	r.onlyInteractive = append(r.onlyInteractive, name)
}

func (rg *registry) RegisterGlobalInitializer(name string, ita Initializer) {
	rg.initters[name] = ita
}

func (rg *registry) RegisterCommand(name string, cmd Commander, opts ...RegisterCommandOption) error {
	if ok := validName(name); !ok {
		return fmt.Errorf("invalid name: %s", name)
	}
	rg.commands[name] = cmd
	for _, opt := range opts {
		opt.Apply(rg, name)
	}
	return nil
}

func (rg *registry) IsAvailable(isInteractive bool, cmdName string) bool {
	for _, o := range rg.onlyInteractive {
		if cmdName == o && !isInteractive {
			return false
		}
	}
	return true
}

func (rg *registry) RegisterAugmentor(name string, ag Augmentor) {
	rg.augmentors[name] = ag
}

func (rg *registry) RegisterPrinter(name string, pr Printer) {
	rg.printers[name] = pr
}

func (rg *registry) GlobalInitializers() RegistryItems[Initializer] {
	return sortedItems(rg.initters)
}

func (rg *registry) Commands() RegistryItems[Commander] {
	return sortedItems(rg.commands)
}

func (rg *registry) Augmentors() RegistryItems[Augmentor] {
	return sortedItems(rg.augmentors)
}

func (rg *registry) Printers() map[string]Printer {
	return rg.printers
}

func (rg *registry) PrinterNames() []string {
	return mk.KeysOf(rg.printers)
}

func sortedItems[T any](d map[string]T) RegistryItems[T] {
	r := make(RegistryItems[T], 0, len(d))
	for name, item := range d {
		r = append(r, RegistryItem[T]{Name: name, Item: item})
	}
	sort.Slice(r, func(i, j int) bool {
		return r[i].Name < r[j].Name
	})
	return r
}

var cmdRegex = MustValue(regexp.Compile(`^[a-z-]+(:[a-z-]+)?$`))

func validName(name string) bool {
	return cmdRegex.Match([]byte(name))
}
