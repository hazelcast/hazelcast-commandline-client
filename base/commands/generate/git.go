package generate

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
)

func cloneTemplates(tsDir string, t string) error {
	err := os.MkdirAll(tsDir, 0700)
	if err != nil {
		return err
	}
	exists, err := templateExists(tsDir, t)
	if err != nil {
		return err
	}
	if !exists {
		//TODO: sparse checkout to be able to download single template instead of cloning the whole template repository
		// it requires git config core.sparseCheckout true
		// and directories should be listed inside .git/info/sparse-checkout
		// https://stackoverflow.com/questions/600079/how-do-i-clone-a-subdirectory-only-of-a-git-repository
		cmd := exec.Command("git", "clone", "--depth", "1", hzTemplatesRepository)
		cmd.Dir = paths.Home()
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			return fmt.Errorf("cloning templates repository (%s): %s", hzTemplatesRepository, stderr.String())
		}
	}
	return nil
}
