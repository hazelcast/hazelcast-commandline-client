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
Generates compact serializer from the given schema and for the given programming language. (BETA)
`
	long := `
Generates compact serializer from the given schema and for the given programming language.
You can use this command to automatically generate compact serializers instead of implementing them.
See: https://docs.hazelcast.com/hazelcast/latest/serialization/compact-serialization#implementing-compactserializer

A schema allows you to:
- describe the contents of a compact class using supported field types
- import other schema
- specify a namespaces for schema files and reference other namespaces
- define cyclic references between classes
- reference classes that are not present in the given schemas

A schema is written in YAML.Schema format is given below:


	namespace: <namespace of the class>
	# note that other schema files can be imported with relative path to this yaml file
	imports:
	 - someOtherSchema.yaml
	# All objects in this file will share the same namespace. 
	classes:
	 - name: <name of the class>
	   fields:
		 - name: <field name>
		   type: <field type>
		   external: bool # to mark external types (external to this yaml file)


- namespace: 	Used for logical grouping of classes. Typically, for every namespace, you will have a schema file. 
		Namespace is optional. If not provided, the classes will be generated at global namespace (no namespace). 
		The user should provide the language specific best practice when using the namespace. 
		The tool will use the namespace while generating code if provided.

- imports: 	Used to import other schema files. 
		The type definitions in the imported yaml schemas can be used within this yaml file. 
		Cyclic imports will be checked and handled properly. 
		For this version of the tool, an import can only be a single file name 
		and the tool will assume all yaml files imported will be in the same directory as the importing schema file.

- classes: 	Used to define classes in the schema
 - name: 	Name of the class
 - fields: 	Fields of the class
  - name: 	Name of the field
  - type: 	Type of the field. 
		Normally you should refer to another class as namespace.classname. 
		You can use a class without namespace when the class is defined in the same schema yaml file. 
		type can be one of the following:
		- boolean
		- boolean[]
		- int8
		- int8[]
		- int16
		- int16[]
		- int32
		- int32[]
		- int64
		- int64[]
		- float32
		- float32[]
		- float64
		- float64[]
		- string
		- string[]
		- date
		- date[]
		- time
		- time[]
		- timestamp
		- timestamp[]
		- timestampWithTimezone
		- timestampWithTimezone[]
		- nullableBoolean
		- nullableBoolean[]
		- nullableInt8
		- nullableInt8[]
		- nullableInt16
		- nullableInt16[]
		- nullableInt32
		- nullableInt32[]
		- nullableInt64
		- nullableInt64[]
		- nullableFloat32
		- nullableFloat32[]
		- nullableFloat64
		- nullableFloat64[]
		- <OtherCompactClass[]>
  - external: 	Used to mark if the type is external. 
		If a field is external, the tool will not check if it is imported and available. 
		External types are managed by the user and not generated by the tool.
		
		The serializer of an external field can be a custom serializer which is hand written, 
		the zero-config serializer for Java and .NET, or previously generated using the tool. 
		This flag will enable such mixed use cases.
		
		In generated code, external types are imported exactly what as the "type" of the field, 
		hence for languages like Java the user should enter the full package name together with the class. 
		E.g. type: com.app1.dto.Address
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
