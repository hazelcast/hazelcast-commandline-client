//go:build base

package serializer

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hazelcast/hazelcast-commandline-client/clc"
	. "github.com/hazelcast/hazelcast-commandline-client/internal/check"
	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

type GenerateCmd struct{}

const (
	flagLanguage  = "language"
	flagOutputDir = "output-dir"
)

func (g GenerateCmd) Init(cc plug.InitContext) error {
	cc.SetCommandUsage("generate [schema] [flags]")
	//TODO: cc.SetCommandHelp(long, short)
	//TODO: must be same with the docs
	cc.AddStringFlag(flagLanguage, "", "", true, "language to generate compact serializer")
	cc.AddStringFlag(flagOutputDir, "", ".", false, "output directory")
	cc.SetPositionalArgCount(1, 1)
	return nil
}

func (g GenerateCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	schemaPath := ec.Args()[0]
	language := ec.Props().GetString(flagLanguage)
	outputDir := ec.Props().GetString(flagOutputDir)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Generating compact serializer from %s schema for %s", schemaPath, language))
		return nil, generateCompactSerializer(schemaPath, language, outputDir)
	})
	if err != nil {
		return err
	}
	stop()
	return nil
}

func generateCompactSerializer(schemaPath, language, outputDir string) error {
	err := validateInputs(schemaPath, language)
	if err != nil {
		return err
	}
	f, err := readSchema(schemaPath)
	if err != nil {
		return err
	}
	sch, err := parseSchema(f)
	if err != nil {
		return err
	}
	schemaDir := filepath.Dir(schemaPath)
	err = processSchema(schemaDir, &sch)
	if err != nil {
		return err
	}
	ccs, err := generateCompactClasses(language, sch)
	if err != nil {
		return err
	}
	err = saveCompactClasses(outputDir, ccs)
	if err != nil {
		return err
	}
	printFurtherToDoInfo(language, ccs)
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("serializer:generate", &GenerateCmd{}))
}
