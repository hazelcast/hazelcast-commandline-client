package plug

import (
	"fmt"
	"sync"
)

type BlockingValue func() (any, error)

type ReadOnlyProperties interface {
	Get(name string) (any, bool)
	GetBlocking(name string) (any, error)
	GetString(name string) string
	GetBool(name string) bool
	GetInt(name string) int64
}

type Properties struct {
	mu *sync.RWMutex
	// properties stack
	pss []map[string]any
	// current properties
	ps  map[string]any
	psb map[string]BlockingValue
}

func NewProperties() *Properties {
	pss := []map[string]any{{}}
	return &Properties{
		mu:  &sync.RWMutex{},
		pss: pss,
		ps:  pss[len(pss)-1],
		psb: map[string]BlockingValue{},
	}
}

func (p *Properties) Push() {
	p.mu.Lock()
	ps := map[string]any{}
	p.pss = append(p.pss, ps)
	p.ps = p.pss[len(p.pss)-1]
	p.mu.Unlock()
}

func (p *Properties) Pop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.pss) == 0 {
		return
	}
	p.pss = p.pss[:len(p.pss)-1]
	p.ps = p.pss[len(p.pss)-1]
}

func (p *Properties) Set(name string, value any) {
	p.mu.Lock()
	p.ps[name] = value
	p.mu.Unlock()
}

func (p *Properties) SetBlocking(name string, value BlockingValue) {
	p.mu.Lock()
	p.psb[name] = value
	p.mu.Unlock()
}

func (p *Properties) Get(name string) (any, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	// traverse the stack to find the name
	for i := len(p.pss) - 1; i >= 0; i-- {
		ps := p.pss[i]
		v, ok := ps[name]
		if ok {
			return v, true
		}
	}
	return nil, false
}

func (p *Properties) GetBlocking(name string) (any, error) {
	p.mu.RLock()
	v, ok := p.psb[name]
	p.mu.RUnlock()
	if !ok {
		// TODO:
		return nil, nil
	}
	return v()
}

func (p *Properties) GetString(name string) string {
	v, ok := p.Get(name)
	if !ok {
		return ""
	}
	if sv, ok := v.(string); ok {
		return sv
	}
	return fmt.Sprintf("%v", v)
}

func (p *Properties) GetBool(name string) bool {
	v, ok := p.Get(name)
	if ok {
		if bv, ok := v.(bool); ok {
			return bv
		}
	}
	return false
}

func (p *Properties) GetInt(name string) int64 {
	v, ok := p.Get(name)
	if ok {
		if bv, ok := v.(int64); ok {
			return bv
		}
	}
	return 0

}

type RegistryItem[T any] struct {
	Name string
	Item T
}

type RegistryItems[T any] []RegistryItem[T]

func (ri RegistryItems[T]) Map() map[string]T {
	m := make(map[string]T, len(ri))
	for _, x := range ri {
		m[x.Name] = x.Item
	}
	return m
}
