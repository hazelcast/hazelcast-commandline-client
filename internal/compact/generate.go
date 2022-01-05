package compact

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"gopkg.in/yaml.v2"
)

type Field struct {
	Name    string `yaml:"name"`
	Type    string `yaml:"type"`
	Default string `yaml:"default"`
}

type Class struct {
	Name   string `yaml:"name"`
	Fields []Field
}
type Schemas struct {
	Classes []Class
}

var builtinTypes = map[string]bool{
	"boolean":               true,
	"int8":                  true,
	"int16":                 true,
	"int32":                 true,
	"int64":                 true,
	"float32":               true,
	"float64":               true,
	"string":                true,
	"decimal":               true,
	"time":                  true,
	"date":                  true,
	"timestamp":             true,
	"timestampWithTimezone": true,
	"nullableBoolean":       true,
	"nullableInt8":          true,
	"nullableInt16":         true,
	"nullableInt32":         true,
	"nullableInt64":         true,
	"nullableFloat32":       true,
	"nullableFloat64":       true,
}

var fixedSizeTypes = map[string]bool{
	"boolean": true,
	"int8":    true,
	"int16":   true,
	"int32":   true,
	"int64":   true,
	"float32": true,
	"float64": true,
}

var languageGenerators = map[string]func(schemas Schemas, namespace string, outputDir string) (string, error){
	"java":   JavaGenerate,
	"cpp":    nil,
	"csharp": nil,
}

func Generate(language string, schemaFilePath string, outputDir string, namespace string) error {
	yamlFile, err := ioutil.ReadFile(schemaFilePath)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	schemas := Schemas{}
	err = yaml.Unmarshal(yamlFile, &schemas)
	if err != nil {
		log.Fatalf("File Format Error: #%v", err)
	}
	err = Validate(schemas)
	if err != nil {
		log.Fatalf("File Format Error: #%v", err)
	}
	generator := languageGenerators[language]
	hint, err := generator(schemas, outputDir, namespace)
	//TODO add silent
	log.Printf(hint)
	return err
}

func Validate(schemas Schemas) error {
	compactsByName := make(map[string]bool)

	// detect duplicate classes and field names
	for _, class := range schemas.Classes {

		fieldsByName := make(map[string]bool)
		for _, field := range class.Fields {
			if _, ok := fieldsByName[field.Name]; ok {
				return errors.New(fmt.Sprintf("a field with name %s already exists in %s", field.Name, class.Name))
			} else {
				fieldsByName[field.Name] = true
			}
		}

		if _, ok := compactsByName[class.Name]; ok {
			return errors.New(fmt.Sprintf("a compact with name %s already exists", class.Name))
		} else {
			compactsByName[class.Name] = true
		}
	}

	// detect if all field types are valid
	for _, class := range schemas.Classes {
		for _, field := range class.Fields {
			componentType := strings.TrimSuffix(field.Type, "[]")
			if _, ok := builtinTypes[componentType]; ok {
				continue
			}
			if _, ok := compactsByName[componentType]; ok {
				continue
			}
			return errors.New(fmt.Sprintf("field type %s is not one of the"+
				" builtin types and or defined", componentType))
		}
	}

	// detect the wrong usage of the default
	for _, class := range schemas.Classes {
		for _, field := range class.Fields {
			if _, ok := fixedSizeTypes[field.Type]; !ok {
				if field.Default != "" {
					return errors.New(fmt.Sprintf("default section is not allowed for field type %s", field.Type))
				}
			}
		}
	}
	return nil
}
