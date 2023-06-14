//go:build base

package generate

import (
	"context"
	"fmt"
	"io/fs"
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
	cc.AddStringFlag(projectTemplate, "", "", true, "name of the template")
	cc.AddStringFlag(projectOutput, "", ".", false, "output directory for the project to be generated")
	cc.SetCommandUsage("project [--template] [flags]")
	help := "Generate a project from template"
	cc.SetCommandHelp(help, help)
	return nil
}

func (g ProjectCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	templateName := ec.Props().GetString(projectTemplate)
	outputDir := ec.Props().GetString(projectOutput)
	templatesDir := paths.Templates()
	err := cloneTemplates(templatesDir, templateName)
	if err != nil {
		return err
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Generating project from template %s", templateName))
		return nil, generateProject(outputDir, ec, templateName)
	})
	stop()
	if err != nil {
		return err
	}
	return nil
}

func generateProject(outputDir string, ec plug.ExecContext, templateName string) error {
	currentTemplateDir := paths.TemplatePath(templateName)
	return filepath.WalkDir(currentTemplateDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			err = os.MkdirAll(filepath.Join(outputDir, d.Name()), 0700)
			if err != nil {
				return err
			}
		} else {
			ext := path.Ext(d.Name())
			// skip files with . and _ prefix unless their extension is ".keep"
			if ext != keepExt && (strings.HasPrefix(d.Name(), hiddenFilePrefix) || strings.HasPrefix(d.Name(), underscorePrefix)) {
				return nil
			}
			if ext == templateExt {
				err = applyTemplateAndCopyToTarget(ec, currentTemplateDir, d.Name(), filepath.Join(outputDir, templateName))
				if err != nil {
					return err
				}
				return nil
			}
			// copy everything else
			err = copyToTarget(currentTemplateDir, d.Name(), filepath.Join(outputDir, templateName))
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func init() {
	Must(plug.Registry.RegisterCommand("generate:project", &ProjectCmd{}))
}
