package project

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport"
)

func cloneTemplate(tsDir string, t string) error {
	u := templateRepoURL(t)
	_, err := git.PlainClone(filepath.Join(tsDir, t), false, &git.CloneOptions{
		URL:      u,
		Progress: nil,
		Depth:    1,
	})
	if err != nil {
		if errors.Is(err, transport.ErrAuthenticationRequired) {
			return fmt.Errorf("repository %s may not exist or requires authentication", u)
		}
		return err
	}
	return nil
}

func templateRepoURL(tName string) string {
	u := os.Getenv(envTemplateSource)
	if u == "" {
		u = hzTemplatesRepository
	}
	u = strings.TrimSuffix(u, "/")
	return fmt.Sprintf("%s/%s", u, tName)
}
