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

	"github.com/hazelcast/hazelcast-commandline-client/base/commands"
	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/cmd"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	"github.com/hazelcast/hazelcast-commandline-client/clc/ux/stage"
	"github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/mk"
	"github.com/hazelcast/hazelcast-commandline-client/internal/output"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
	iserialization "github.com/hazelcast/hazelcast-commandline-client/internal/serialization"
)

const (
	argTemplateName      = "templateName"
	argTitleTemplateName = "template name"
	argPlaceholder       = "placeholder"
	argTitlePlaceholder  = "placeholder"
)

var regexpValidKey = regexp.MustCompile(`^[a-z0-9_]+$`)

type CreateCommand struct{}

func (pc CreateCommand) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("create")
	short := "Create project from the given template (BETA)"
	long := longHelp()
	cc.SetCommandHelp(long, short)
	cc.AddStringFlag(commands.FlagOutputDir, "o", "", false, "the directory to create the project at")
	cc.AddStringArg(argTemplateName, argTitleTemplateName)
	cc.AddKeyValueSliceArg(argPlaceholder, argTitlePlaceholder, 0, clc.MaxArgs)
	return nil
}

func (pc CreateCommand) Exec(ctx context.Context, ec plug.ExecContext) error {
	cmd.IncrementMetric(ctx, ec, "total.project")
	templateName := ec.GetStringArg(argTemplateName)
	outputDir := ec.Props().GetString(commands.FlagOutputDir)
	if outputDir == "" {
		outputDir = templateName
	}
	var stages []stage.Stage[any]
	templatesDir := paths.Templates()
	templateExists := paths.Exists(filepath.Join(templatesDir, templateName))
	if templateExists {
		stages = append(stages, stage.Stage[any]{
			ProgressMsg: "Updating the template",
			SuccessMsg:  fmt.Sprintf("Updated template '%s'", templateName),
			FailureMsg:  "Failed updating the template",
			Func: func(ctx context.Context, status stage.Statuser[any]) (any, error) {
				err := updateTemplate(ctx, templatesDir, templateName)
				if err != nil {
					ec.Logger().Error(err)
					return nil, stage.IgnoreError(err)
				}
				return nil, nil
			},
		})
	} else {
		stages = append(stages, stage.Stage[any]{
			ProgressMsg: "Retrieving the template",
			SuccessMsg:  fmt.Sprintf("Retrieved template '%s'", templateName),
			FailureMsg:  "Failed retrieving the template",
			Func: func(ctx context.Context, status stage.Statuser[any]) (any, error) {
				ec.Logger().Debug(func() string {
					return fmt.Sprintf("template %s does not exist, cloning it into %s", templateName, templatesDir)
				})
				err := cloneTemplate(ctx, templatesDir, templateName)
				if err != nil {
					ec.Logger().Error(err)
					return nil, err
				}
				return nil, nil
			},
		})
	}
	stages = append(stages, stage.Stage[any]{
		ProgressMsg: "Creating the project",
		SuccessMsg:  "Created the project",
		FailureMsg:  "Failed creating the project",
		Func: func(ctx context.Context, status stage.Statuser[any]) (any, error) {
			return nil, createProject(ec, outputDir, templateName)
		},
	})
	_, err := stage.Execute[any](ctx, ec, nil, stage.NewFixedProvider(stages...))
	if err != nil {
		return err
	}
	ec.PrintlnUnnecessary("")
	return ec.AddOutputRows(ctx, output.Row{
		output.Column{
			Name:  "Path",
			Type:  iserialization.TypeString,
			Value: outputDir,
		},
	})
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
	check.Must(plug.Registry.RegisterCommand("project:create", &CreateCommand{}))
}
