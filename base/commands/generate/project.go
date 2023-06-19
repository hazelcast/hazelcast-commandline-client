//go:build base

package generate

import (
	"context"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type ProjectCmd struct{}

func (g ProjectCmd) Init(cc plug.InitContext) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	_, f := filepath.Split(wd)
	cc.AddStringFlag(projectOutput, "o", wd, false, "output directory for the project to be generated")
	cc.AddStringFlag(projectName, "", f, false, "name of the created project")
	cc.SetPositionalArgCount(1, math.MaxInt)
	cc.SetCommandUsage("project [template-name] [flags]")
	help := "Generate a project from template"
	cc.SetCommandHelp(help, help)
	return nil
}

func (g ProjectCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	templateName := ec.Args()[0]
	outputDir := ec.Props().GetString(projectOutput)
	pName := ec.Props().GetString(projectName)
	templatesDir := paths.Templates()
	templateExists := paths.Exists(filepath.Join(templatesDir, templateName))
	if !templateExists {
		err := cloneTemplate(templatesDir, templateName)
		if err != nil {
			return err
		}
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Generating project from template %s", templateName))
		return nil, createProject(ec, outputDir, templateName, pName)
	})
	stop()
	if err != nil {
		return err
	}
	return nil
}

func createProject(ec plug.ExecContext, outputDir, templateName, pName string) error {
	sourceDir := paths.TemplatePath(templateName)
	return filepath.WalkDir(sourceDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(outputDir, pName, strings.Split(p, templateName)[1])
		if d.IsDir() {
			if isSkip(d) {
				// skip dir and its subdirectories and files
				return filepath.SkipDir
			}
			err = os.MkdirAll(target, 0700)
			if err != nil {
				return err
			}
		} else {
			if isSkip(d) {
				// skip only current file
				return nil
			}
			if path.Ext(d.Name()) == templateExt {
				err = applyTemplateAndCopyToTarget(ec, sourceDir, p, target)
				if err != nil {
					return err
				}
				return nil
			}
			// copy everything else
			err = copyToTarget(p, target, path.Ext(d.Name()) == ".keep")
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func isSkip(d fs.DirEntry) bool {
	if (isHidden(d) && !hasKeepExt(d)) || isDefaultPropertiesFile(d) {
		return true
	}
	return false
}

func isHidden(d fs.DirEntry) bool {
	return strings.HasPrefix(d.Name(), hiddenFilePrefix) || strings.HasPrefix(d.Name(), underscorePrefix)
}

func hasKeepExt(d fs.DirEntry) bool {
	return path.Ext(d.Name()) == keepExt
}

func isDefaultPropertiesFile(d fs.DirEntry) bool {
	return d.Name() == defaultPropertiesFileName
}

func init() {
	Must(plug.Registry.RegisterCommand("generate:project", &ProjectCmd{}))
}
