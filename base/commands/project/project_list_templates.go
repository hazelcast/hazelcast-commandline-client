package project

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/clc/store"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/log"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	"github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

type ListCmd struct{}

const (
	flagRefresh          = "refresh"
	flagLocal            = "local"
	nextFetchTimeKey     = "project.templates.nextFetchTime"
	templatesKey         = "project.templates"
	cacheRefreshInterval = 10 * time.Minute
)

type Template struct {
	Name   string `json:"name"`
	Source string
}

func (lc ListCmd) Init(cc plug.InitContext) error {
	cc.SetPositionalArgCount(0, 0)
	cc.SetCommandUsage("list-templates [flags]")
	cc.AddBoolFlag(flagRefresh, "", false, false, "fetch most recent templates from remote")
	cc.AddBoolFlag(flagLocal, "", false, false, "list the templates which exist on local environment")
	help := "Lists templates that can be used while creating projects."
	cc.SetCommandHelp(help, help)
	return nil
}

func (lc ListCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	isLocal := ec.Props().GetBool(flagLocal)
	isRefresh := ec.Props().GetBool(flagRefresh)
	if isLocal && isRefresh {
		return fmt.Errorf("%s and %s flags are mutually exclusive", flagRefresh, flagLocal)
	}
	ts, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Listing templates"))
		return listTemplates(ec.Logger(), isLocal, isRefresh)
	})
	if err != nil {
		return err
	}
	stop()
	tss := ts.([]Template)
	if len(tss) == 0 {
		ec.PrintlnUnnecessary("No templates found")
	}
	rows := make([]output.Row, len(tss))
	for i, t := range tss {
		rows[i] = output.Row{
			output.Column{
				Name:  "Source",
				Type:  serialization.TypeString,
				Value: t.Source,
			},
			output.Column{
				Name:  "Name",
				Type:  serialization.TypeString,
				Value: t.Name,
			},
		}
	}
	return ec.AddOutputRows(ctx, rows...)
}

func listTemplates(logger log.Logger, isLocal bool, isRefresh bool) ([]Template, error) {
	sa := store.NewStoreAccessor(filepath.Join(paths.Caches(), "templates"), logger)
	if isLocal {
		return listLocalTemplates()
	}
	var fetch bool
	var err error
	if fetch, err = shouldFetch(sa); err != nil {
		logger.Debugf("Error: checking template list expiry: %w", err)
		// there is an error with database, so fetch templates from remote
		fetch = true
	}
	if fetch || isRefresh {
		ts, err := fetchTemplates()
		if err != nil {
			return nil, err
		}
		err = updateCache(sa, ts)
		if err != nil {
			logger.Debugf("Error: Updating templates cache: %w", err)
		}
	}
	return listFromCache(sa)
}

func listLocalTemplates() ([]Template, error) {
	var templates []Template
	ts, err := paths.FindAll(paths.Templates(), func(basePath string, entry os.DirEntry) (ok bool) {
		return entry.IsDir()
	})
	if err != nil {
		return nil, err
	}
	for _, t := range ts {
		templates = append(templates, Template{Name: t, Source: "local"})
	}
	return templates, nil
}

func fetchTemplates() ([]Template, error) {
	var templates []Template
	resp, err := http.Get(makeRepositoriesURL())
	if err != nil {
		return nil, err
	}
	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var data []map[string]any
	err = json.Unmarshal(respData, &data)
	if err != nil {
		return nil, err
	}
	for _, d := range data {
		var tName string
		var ok bool
		if tName, ok = d["full_name"].(string); !ok {
			return nil, errors.New("error fetching repositories in the organization")
		}
		sName := strings.Split(tName, "/")
		source := fmt.Sprintf("%s/%s", "github.com", sName[0])
		name := sName[1]
		templates = append(templates, Template{Name: name, Source: source})
	}
	return templates, nil
}

func updateNextFetchTime(sa *store.StoreAccessor) error {
	_, err := sa.WithLock(func(s *store.Store) (any, error) {
		v := []byte(strconv.FormatInt(time.Now().Add(cacheRefreshInterval).Unix(), 10))
		return nil, s.SetEntry([]byte(nextFetchTimeKey), v)
	})
	return err
}

func makeRepositoriesURL() string {
	s := strings.TrimPrefix(templateOrgURL(), "https://github.com/")
	ss := strings.ReplaceAll(s, "/", "")
	return fmt.Sprintf("https://api.github.com/users/%s/repos", ss)
}

func shouldFetch(s *store.StoreAccessor) (bool, error) {
	entry, err := s.WithLock(func(s *store.Store) (any, error) {
		return s.GetEntry([]byte(nextFetchTimeKey))
	})
	if err != nil {
		if errors.Is(err, store.ErrKeyNotFound) {
			return true, nil
		}
		return false, err
	}
	var fetchTS time.Time
	t, err := strconv.ParseInt(string(entry.([]byte)), 10, 64)
	if err != nil {
		return false, err
	}
	fetchTS = time.Unix(t, 0)
	if time.Now().After(fetchTS) {
		return true, nil
	}
	return false, nil
}

func updateCache(sa *store.StoreAccessor, templates []Template) error {
	b, err := json.Marshal(templates)
	if err != nil {
		return err
	}
	_, err = sa.WithLock(func(s *store.Store) (any, error) {
		err = s.SetEntry([]byte(templatesKey), b)
		if err != nil {
			return nil, err
		}
		if err = updateNextFetchTime(sa); err != nil {
			return nil, err
		}
		return nil, nil
	})
	return err
}

func listFromCache(sa *store.StoreAccessor) ([]Template, error) {
	var templates []Template
	b, err := sa.WithLock(func(s *store.Store) (any, error) {
		return s.GetEntry([]byte(templatesKey))
	})
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(b.([]byte), &templates)
	if err != nil {
		return nil, err
	}
	return templates, nil
}

func init() {
	Must(plug.Registry.RegisterCommand("project:list-templates", &ListCmd{}))
}
