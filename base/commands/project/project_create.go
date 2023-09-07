//go:build std || project

package project

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/mk"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const (
	argTemplateName      = "templateName"
	argTitleTemplateName = "template name"
	argPlaceholder       = "placeholder"
	argTitlePlaceholder  = "placeholder"
)

var regexpValidKey = regexp.MustCompile(`^[a-z0-9_]+$`)

type CreateCmd struct{}

func (pc CreateCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("create")
	short := "Create project from the given template (BETA)"
	long := longHelp()
	cc.SetCommandHelp(long, short)
	cc.AddStringFlag(flagOutputDir, "o", "", false, "the directory to create the project at")
	cc.AddStringArg(argTemplateName, argTitleTemplateName)
	cc.AddKeyValueSliceArg(argPlaceholder, argTitlePlaceholder, 0, clc.MaxArgs)
	return nil
}

func (pc CreateCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	templateName := ec.GetStringArg(argTemplateName)
	outputDir := ec.Props().GetString(flagOutputDir)
	if outputDir == "" {
		outputDir = templateName
	}
	templatesDir := paths.Templates()
	templateExists := paths.Exists(filepath.Join(templatesDir, templateName))
	if !templateExists {
		ec.Logger().Debug(func() string {
			return fmt.Sprintf("template %s does not exist, cloning it into %s", templateName, templatesDir)
		})
		err := cloneTemplate(templatesDir, templateName)
		if err != nil {
			ec.Logger().Error(err)
			return err
		}
	}
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Creating project from template %s", templateName))
		return nil, createProject(ec, outputDir, templateName)
	})
	stop()
	if err != nil {
		return err
	}
	return nil
}

func createProject(ec plug.ExecContext, outputDir, templateName string) error {
	sourceDir := paths.ResolveTemplatePath(templateName)
	vs, err := loadValues(ec, sourceDir)
	if err != nil {
		return err
	}
	ec.Logger().Debug(func() string {
		return fmt.Sprintf("available placeholders: %+v", mk.KeysOf(vs))
	})
	err = filepath.WalkDir(sourceDir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(outputDir, strings.Split(path, templateName)[1])
		if entry.IsDir() {
			if isSkip(entry) {
				// skip dir and its subdirectories and files
				return filepath.SkipDir
			}
			err = os.MkdirAll(target, 0700)
			if err != nil {
				return err
			}
		} else {
			if isSkip(entry) {
				// skip only current file
				return nil
			}
			if hasTemplateExt(entry) {
				err = applyTemplateAndCopyToTarget(vs, path, target)
				if err != nil {
					return err
				}
				return nil
			}
			// copy everything else
			err = copyToTarget(path, target, hasKeepExt(entry))
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	ec.Logger().Info(fmt.Sprintf("Successfully created project: %s", outputDir))
	return nil
}

func loadValues(ec plug.ExecContext, sourceDir string) (map[string]string, error) {
	vs, err := loadFromDefaults(sourceDir)
	if err != nil {
		return nil, err
	}
	loadFromProps(ec, vs)
	if err = updatePropsWithUserValues(ec, vs); err != nil {
		return nil, err
	}
	return vs, nil
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
	return filepath.Ext(d.Name()) == keepExt
}

func hasTemplateExt(d fs.DirEntry) bool {
	return filepath.Ext(d.Name()) == templateExt
}

func isDefaultPropertiesFile(d fs.DirEntry) bool {
	return d.Name() == defaultsFileName
}

func init() {
	Must(plug.Registry.RegisterCommand("project:create", &CreateCmd{}))
}
