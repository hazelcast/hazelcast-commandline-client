//go:build std || project

package project

import (
	"io"
	"os"
	"text/template"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
)

func applyTemplateAndCopyToTarget(vars map[string]string, source, dest string) error {
	base, _ := paths.SplitExt(dest)
	f, err := os.Create(base)
	if err != nil {
		return err
	}
	defer f.Close()
	t, err := template.ParseFiles(source)
	if err != nil {
		return err
	}
	if err = t.Execute(f, vars); err != nil {
		return err
	}
	return nil
}

func copyToTarget(source, dest string, removeExt bool) error {
	sf, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sf.Close()
	if removeExt {
		dest, _ = paths.SplitExt(dest)
	}
	df, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer df.Close()
	if _, err = io.Copy(df, sf); err != nil {
		return err
	}
	return nil
}
