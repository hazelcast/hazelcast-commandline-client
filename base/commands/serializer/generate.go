//go:build std || serializer

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
	short := `
Generates compact serializer from the given schema and for the given programming language.
`
	long := `
Generates compact serializer from the given schema and for the given programming language.
You can use this command to automatically generate compact serializers instead of implementing them.
See: https://docs.hazelcast.com/hazelcast/5.3/serialization/compact-serialization#implementing-compactserializer
`
	cc.SetCommandHelp(long, short)
	cc.AddStringFlag(flagLanguage, "l", "", true, "programming language to use for the generated code")
	cc.AddStringFlag(flagOutputDir, "o", ".", false, "output directory for the generated files")
	cc.SetPositionalArgCount(1, 1)
	return nil
}

func (g GenerateCmd) Exec(ctx context.Context, ec plug.ExecContext) error {
	schemaPath := ec.Args()[0]
	language := ec.Props().GetString(flagLanguage)
	outputDir := ec.Props().GetString(flagOutputDir)
	_, stop, err := ec.ExecuteBlocking(ctx, func(ctx context.Context, sp clc.Spinner) (any, error) {
		sp.SetText(fmt.Sprintf("Generating compact serializer for %s", language))
		return nil, generateCompactSerializer(ec, schemaPath, language, outputDir)
	})
	if err != nil {
		return err
	}
	stop()
	return nil
}

func generateCompactSerializer(ec plug.ExecContext, schemaPath, language, outputDir string) error {
	err := validateInputs(schemaPath, language)
	if err != nil {
		return fmt.Errorf("validating inputs: %w", err)
	}
	f, err := readSchema(schemaPath)
	if err != nil {
		return fmt.Errorf("reading the schema: %w", err)
	}
	sch, err := parseSchema(f)
	if err != nil {
		return fmt.Errorf("parsing the schema: %w", err)
	}
	schemaDir := filepath.Dir(schemaPath)
	err = processSchema(schemaDir, &sch)
	if err != nil {
		return fmt.Errorf("processing the schema: %w", err)
	}
	ccs, err := generateCompactClasses(language, sch)
	if err != nil {
		return fmt.Errorf("generating compact classes: %w", err)
	}
	err = saveCompactClasses(outputDir, ccs)
	if err != nil {
		return fmt.Errorf("saving compact classes: %w", err)
	}
	printFurtherToDoInfo(ec, language, ccs)
	return nil
}

func init() {
	Must(plug.Registry.RegisterCommand("serializer:generate", &GenerateCmd{}))
}
