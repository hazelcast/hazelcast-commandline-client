package config

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/hazelcast/hazelcast-go-client"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/errors"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type Provider interface {
	GetString(key string) string
	Set(key string, value any)
	All() map[string]any
	BindFlag(name string, flag *pflag.Flag)
	ClientConfig(ctx context.Context, ec plug.ExecContext) (hazelcast.Config, error)
}

type FileProvider struct {
	mu                 *sync.RWMutex
	cfg                map[string]any
	defaults           map[string]any
	keys               map[string]struct{}
	boundFlags         map[string]*pflag.Flag
	canUseClientConfig bool
	hasClientConfig    bool
	clientCfg          hazelcast.Config
	path               string
}

func NewFileProvider(path string) (*FileProvider, error) {
	p := &FileProvider{
		mu:         &sync.RWMutex{},
		cfg:        map[string]any{},
		keys:       map[string]struct{}{},
		defaults:   map[string]any{},
		boundFlags: map[string]*pflag.Flag{},
	}
	if err := p.load(path); err != nil {
		return nil, err
	}
	return p, nil
}

func (p *FileProvider) load(path string) error {
	path = paths.ResolveConfigPath(path)
	if path == "" {
		// there is no default config, user will be prompted for config later
		return nil
	}
	if !paths.Exists(path) {
		return fmt.Errorf("configuration does not exist %s: %w", path, os.ErrNotExist)
	}
	p.path = path
	p.keys[path] = struct{}{}
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading configuration: %w", err)
	}
	m := map[string]any{}
	if err := yaml.Unmarshal(b, m); err != nil {
		return fmt.Errorf("loading configuration: %w", err)
	}
	p.traverseMap("", m)
	p.canUseClientConfig = true
	return nil
}

func (p *FileProvider) GetString(key string) string {
	if key == clc.PropertyConfig {
		return p.path
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	f, ok := p.boundFlags[key]
	if ok && f.Changed {
		return f.Value.String()
	}
	v, ok := p.get(key)
	if ok {
		return v.(string)
	}
	return ""
}

func (p *FileProvider) Set(key string, value any) {
	p.mu.Lock()
	p.defaults[key] = value
	p.keys[key] = struct{}{}
	p.mu.Unlock()
}

func (p *FileProvider) All() map[string]any {
	p.mu.RLock()
	m := make(map[string]any, len(p.cfg))
	for k := range p.keys {
		v, ok := p.get(k)
		if ok {
			m[k] = v
		}
	}
	p.mu.RUnlock()
	return m
}

func (p *FileProvider) BindFlag(name string, flag *pflag.Flag) {
	p.mu.Lock()
	p.boundFlags[name] = flag
	p.mu.Unlock()
}

func (p *FileProvider) ClientConfig(ctx context.Context, ec plug.ExecContext) (hazelcast.Config, error) {
	cc, ok := p.clientConfig()
	if ok {
		return cc, nil
	}
	if !p.canUseClientConfig {
		return hazelcast.Config{}, errors.ErrNoClusterConfig
	}
	cc, err := MakeHzConfig(p, ec.Logger())
	if err != nil {
		return hazelcast.Config{}, err
	}
	p.mu.Lock()
	p.clientCfg = cc
	p.hasClientConfig = true
	p.mu.Unlock()
	return cc, nil
}

func (p *FileProvider) Get(name string) (any, bool) {
	// XXX: boundFlags is not checked
	p.mu.RLock()
	v, ok := p.get(name)
	p.mu.RUnlock()
	return v, ok
}

func (p *FileProvider) get(name string) (any, bool) {
	// XXX: boundFlags is not checked
	v, ok := p.cfg[name]
	if !ok {
		v, ok = p.defaults[name]
	}
	return v, ok
}

func (p *FileProvider) GetBlocking(name string) (any, error) {
	panic("config.FileProvider: not implemented")
}

func (p *FileProvider) GetBool(name string) bool {
	// XXX: boundFlags is not checked
	v, ok := p.Get(name)
	if !ok {
		return false
	}
	return v.(bool)
}

func (p *FileProvider) GetInt(name string) int64 {
	// XXX: boundFlags is not checked
	v, ok := p.Get(name)
	if !ok {
		return 0
	}
	return v.(int64)
}

func (p *FileProvider) clientConfig() (hazelcast.Config, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.hasClientConfig {
		return p.clientCfg, true
	}
	return hazelcast.Config{}, false
}

func (p *FileProvider) traverseMap(root string, m map[string]any) {
	for ks, v := range m {
		var r string
		if root == "" {
			r = ks
		} else {
			r = strings.Join([]string{root, ks}, ".")
		}
		if mm, ok := v.(map[string]any); ok {
			p.traverseMap(r, mm)
			continue
		}
		p.cfg[r] = v
		p.keys[r] = struct{}{}
	}
}
