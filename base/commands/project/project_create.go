//go:build base

package project

import (
	"context"
	"fmt"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	"github.com/hazelcast/hazelcast-commandline-client/clc/paths"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

const puncPattern = "[[:punct:]]+"

var puncReg = regexp.MustCompile(puncPattern)

type CreateCmd struct{}

func (pc CreateCmd) Init(cc plug.InitContext) error {
	cc.SetPositionalArgCount(2, math.MaxInt)
	cc.SetCommandUsage("create [template-name] [output-dir] [flags]")
	short := "(Beta) Create project from the given template"
	long := fmt.Sprintf(` (Beta) Create project from the given template and project will be created to the given output-dir.
	
Templates are located in the %s organization by default.
You can override it by using CLC_EXPERIMENTAL_TEMPLATE_SOURCE environment variable.

Rules while creating your own templates:

	* Templates are in Go template format.
	  See: https://pkg.go.dev/text/template
	* You can create a "defaults.yaml" file for default values in template's root directory.
	* Template files must have the ".template" extension.
	* Files with "." and "_" prefixes are ignored by default.
	 If want to keep them you must add ".keep" extension to them.
	* Other files are copied verbatim.

Properties are read from the following resources in order:

	1. defaults.yaml (keys cannot contain punctuation)
	2. config.yaml
	3. User passed key-values (keys cannot contain punctuation)

You can use the placeholders in "defaults.yaml" and the following configuration item placeholders:

	* ClusterName
	* ClusterAddress
	* ClusterUser
	* ClusterPassword
	* ClusterDiscoveryToken
	* SslEnabled
	* SslServer
	* SslSkipVerify
	* SslCaPath
	* SslKeyPath
	* SslKeyPassword
	* LogPath
	* LogLevel

Example (Linux and MacOS):

$ export CLC_EXPERIMENTAL_TEMPLATE_SOURCE=https://github.com/my-template-organization
$ clc project create \
	simple-streaming-pipeline-template\
	my-project\
	MyKey1=MyValue1 MyKey2=MyValue2

Example (Windows):


> set CLC_EXPERIMENTAL_TEMPLATE_SOURCE=https://github.com/my-template-organization
> clc project create^
	simple-streaming-pipeline-template^
	my-project^
	MyKey1=MyValue1 MyKey2=MyValue2
`, hzTemplatesRepository)
	cc.SetCommandHelp(long, short)
	return nil
}

func (pc CreateCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	templateName := ec.Args()[0]
	outputDir := ec.Args()[1]
	templatesDir := paths.Templates()
	templateExists := paths.Exists(filepath.Join(templatesDir, templateName))
	if !templateExists {
		ec.Logger().Info(fmt.Sprintf("template %s does not exist, cloning it into %s", templateName, templatesDir))
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
	sourceDir := paths.TemplatePath(templateName)
	vars, err := loadVars(ec, sourceDir)
	if err != nil {
		return err
	}
	ec.Logger().Info(fmt.Sprintf("available placeholders: %+v", reflect.ValueOf(vars).MapKeys()))
	err = filepath.WalkDir(sourceDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		target := filepath.Join(outputDir, strings.Split(p, templateName)[1])
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
			if hasTemplateExt(d) {
				err = applyTemplateAndCopyToTarget(vars, p, target)
				if err != nil {
					return err
				}
				return nil
			}
			// copy everything else
			err = copyToTarget(p, target, hasKeepExt(d))
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

func loadVars(ec plug.ExecContext, sourceDir string) (map[string]string, error) {
	vars := make(map[string]string)
	err := loadFromDefaults(sourceDir, &vars)
	if err != nil {
		return nil, err
	}
	loadFromProps(ec, &vars)
	err = loadFromUserInput(ec, &vars)
	if err != nil {
		return nil, err
	}
	return vars, nil
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
