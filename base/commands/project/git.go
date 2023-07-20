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

func cloneTemplate(baseDir string, name string) error {
	u := templateRepoURL(name)
	_, err := git.PlainClone(filepath.Join(baseDir, name), false, &git.CloneOptions{
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

func templateRepoURL(templateName string) string {
	u := os.Getenv(envTemplateSource)
	if u == "" {
		u = hzTemplatesOrganization
	}
	u = strings.TrimSuffix(u, "/")
	return fmt.Sprintf("%s/%s", u, templateName)
}
