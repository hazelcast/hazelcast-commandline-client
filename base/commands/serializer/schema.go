//go:build std || serializer

package serializer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Schema struct {
	Imports    []string
	Namespace  string
	Classes    []Class
	ClassNames map[string]Class
}

type Class struct {
	Name      string
	Fields    []Field
	Namespace string
}

type Field struct {
	Name     string
	Type     string
	External bool
}

type ClassInfo struct {
	FileName  string
	ClassName string
	Namespace string
}

func readSchema(path string) ([]byte, error) {
	s, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("can not read schema %s", err.Error())
	}
	return s, nil
}

func parseSchema(schema []byte) (Schema, error) {
	schemaMap, err := ConvertYAMLToMap(schema)
	if err != nil {
		return Schema{}, fmt.Errorf("converting YAML to schema: %w", err)
	}
	isValid, schemaErrors, err := validateWithJSONSchema(schemaMap)
	if err != nil {
		return Schema{}, fmt.Errorf("validating the schema: %w", err)
	}
	if !isValid {
		return Schema{}, fmt.Errorf("schema is not valid, validation errors:\n%s", strings.Join(schemaErrors, "\n"))
	}
	return transcode(schemaMap)
}

func transcode(in any) (Schema, error) {
	var schema Schema
	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(in); err != nil {
		return Schema{}, err
	}
	if err := json.NewDecoder(buf).Decode(&schema); err != nil {
		return Schema{}, err
	}
	return schema, nil
}

func processSchema(schemaPath string, schema *Schema) error {
	schema.ClassNames = make(map[string]Class)
	err := registerClasses(*schema, schema.ClassNames)
	if err != nil {
		return err
	}
	err = processImports(*schema, schemaPath, schema.Imports, map[string]struct{}{schemaPath: {}})
	if err != nil {
		return err
	}
	err = checkFieldTypes(*schema, schema.ClassNames)
	if err != nil {
		return err
	}
	return nil
}

func processImports(mainSchema Schema, schemaDir string, importPaths []string, importedPaths map[string]struct{}) error {
	for _, p := range importPaths {
		if err := processImport(mainSchema, schemaDir, p, importedPaths); err != nil {
			return err
		}
	}
	return nil
}

// mainSchema is the main schema
// baseDir is the directory of the schema that imports another schema
// importPath is the path of the schema that is being imported relative to the baseDir
// importedPaths is a map of all the paths that are already imported
func processImport(mainSchema Schema, baseDir string, importPath string, importedPaths map[string]struct{}) error {
	// If the file is already imported, skip it
	newSchemaPath := filepath.Join(baseDir, importPath)
	newSchemaDir := filepath.Dir(newSchemaPath)
	if _, ok := importedPaths[newSchemaPath]; ok {
		return nil
	}
	// We are processing the new import now, so add it to the imported paths
	importedPaths[newSchemaPath] = struct{}{}
	yamlSchema, err := os.ReadFile(newSchemaPath)
	if err != nil {
		return err
	}
	sch, err := parseSchema(yamlSchema)
	if err != nil {
		return err
	}
	err = registerClasses(sch, mainSchema.ClassNames)
	if err != nil {
		return err
	}
	err = processImports(mainSchema, newSchemaDir, sch.Imports, importedPaths)
	if err != nil {
		return err
	}
	err = checkFieldTypes(sch, mainSchema.ClassNames)
	if err != nil {
		return err
	}
	return nil
}

func registerClasses(schema Schema, classNames map[string]Class) error {
	for i := range schema.Classes {
		cls := schema.Classes[i]
		fullName := getClassFullName(cls.Name, schema.Namespace)
		if _, ok := classNames[fullName]; ok {
			return fmt.Errorf("class defined more than once. Compact class with name %s and namespace %s already exist", cls.Name, schema.Namespace)
		}
		cls.Namespace = schema.Namespace
		classNames[fullName] = cls
	}
	return nil
}

func checkFieldTypes(schema Schema, classNames map[string]Class) error {
	for _, c := range schema.Classes {
		fieldNames := make(map[string]struct{}, len(c.Fields))
		for _, f := range c.Fields {
			if _, ok := fieldNames[f.Name]; ok {
				return fmt.Errorf("validation error: '%s' field is defined more than once in class '%s'", f.Name, c.Name)
			}
			typ := f.Type
			fieldNames[f.Name] = struct{}{}
			// if type is an array type, loose the brackets and validate underlying type
			typ = strings.TrimSuffix(typ, "[]")
			if isBuiltInType(typ) {
				continue
			}
			// if field is external, we don't need to validate it
			if f.External {
				continue
			}
			// check if type is a class name
			if isCompactName(schema.Namespace, typ, classNames) {
				continue
			}
			return fmt.Errorf("validation error: field type '%s' is not one of the builtin types or not defined", typ)
		}
	}
	return nil
}

func getClassFullName(className string, namespace string) string {
	if namespace == "" {
		return className
	}
	return namespace + "." + className
}
