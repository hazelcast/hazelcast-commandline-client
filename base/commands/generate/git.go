package generate

import (
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
)

func cloneTemplate(tsDir string, t string) error {
	_, err := git.PlainClone(filepath.Join(tsDir, t), false, &git.CloneOptions{
		URL:      hzTemplatesRepository + t,
		Progress: os.Stdout,
		Depth:    1,
	})
	if err != nil {
		return err
	}
	return nil
}
