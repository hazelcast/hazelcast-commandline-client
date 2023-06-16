package generate

import (
	"os"

	"github.com/go-git/go-git/v5"
)

func cloneTemplate(tsDir string, t string) error {
	repo, err := git.PlainClone(tsDir, false, &git.CloneOptions{
		URL:        hzTemplatesRepository,
		NoCheckout: true,
		Progress:   os.Stdout,
		Depth:      1,
	})
	repo.Config()
	if err != nil {
		return err
	}
	w, err := repo.Worktree()
	options := git.CheckoutOptions{
		SparseCheckoutDirectories: []string{t},
		Branch:                    "refs/heads/main",
	}
	err = w.Checkout(&options)
	if err != nil {
		return err
	}
	return nil
}
