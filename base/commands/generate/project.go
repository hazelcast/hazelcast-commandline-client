//go:build base

package generate

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ProjectCmd struct{}

func (g ProjectCmd) Init(cc plug.InitContext) error {
	cc.AddStringFlag(projectTemplate, "", "", true, "name of the template")
	cc.AddStringFlag(projectOutput, "", ".", false, "output directory for the project to be generated")
	cc.SetCommandUsage("project [--template] [flags]")
	help := "Generate a project from template"
	cc.SetCommandHelp(help, help)
	return nil
}

func (g ProjectCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	t := ec.Props().GetString(projectTemplate)
	o := ec.Props().GetString(projectOutput)
	tsDir := paths.Templates()
	exists, err := templateExists(tsDir, t)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("template %s does not exist", t)
	}
	tDir := paths.TemplatePath(t)
	err = filepath.WalkDir(tDir, func(p string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			err = os.MkdirAll(path.Join(o, d.Name()), 0700)
			if err != nil {
				return err
			}
		} else {
			ext := path.Ext(d.Name())
			// skip files with . and _ prefix unless their extension is ".keep"
			if ext != ".keep" && (strings.HasPrefix(d.Name(), ".") || strings.HasPrefix(d.Name(), "_")) {
				return nil
			}
			if ext == ".template" {
				err = applyTemplateAndCopyToTarget(tDir, d.Name(), path.Join(o, t))
				if err != nil {
					return err
				}
				return nil
			}
			// copy everything else
			err = copyToTarget(tDir, d.Name(), path.Join(o, t))
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("impossible to walk directories: %s", err)
	}
	return nil
}

func applyTemplateAndCopyToTarget(tDir, fileName, destDir string) error {
	destFile, err := os.Create(path.Join(destDir, fileName))
	if err != nil {
		log.Fatal(err)
	}
	defer destFile.Close()
	tmpl, err := template.ParseFiles(path.Join(tDir, fileName))
	if err != nil {
		return err
	}
	data := map[string]string{
		"myVar": "I am a var",
	}
	/* TODO:
	Property sources in the least to most priority:
		default.properties
		Props object
		User passed key-values
	*/
	err = tmpl.Execute(destFile, data)
	if err != nil {
		return err
	}
	return nil
}

func templateExists(tDir, t string) (bool, error) {
	if _, err := os.Stat(filepath.Join(tDir, t)); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func copyToTarget(tDir, fileName, destinationDir string) error {
	sourceFile, err := os.Open(path.Join(tDir, fileName))
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	destinationFile, err := os.Create(path.Join(destinationDir, fileName))
	if err != nil {
		return err
	}
	defer destinationFile.Close()
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("generate:project", &ProjectCmd{}))
}
