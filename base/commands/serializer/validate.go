package serializer

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

func validateInputs(schema, language string) error {
	if !isLangSupported(language) {
		return fmt.Errorf("unsupported language, you can provide one of %s", strings.Join(supportedLanguages, ","))
	}
	if !isSchemaExists(schema) {
		return fmt.Errorf("schema %s does not exist")
	}
	return nil
}

func isLangSupported(lang string) bool {
	lang = strings.ToLower(lang)
	for _, sl := range supportedLanguages {
		if lang == sl {
			return true
		}
	}
	return false
}

func isSchemaExists(schema string) bool {
	//TODO: how to check schema existence, where do the schemas
	return true
}

func validateWithJSONSchema(schema map[string]interface{}) (isValid bool, schemaErrors []string, err error) {
	jsonSchemaString, err := json.Marshal(schema)
	if err != nil {
		return false, nil, err
	}
	return validateJSONSchemaString(string(jsonSchemaString))
}

func validateJSONSchemaString(schema string) (isValid bool, schemaErrors []string, err error) {
	// The json that is validated is called document. We want to validate our compact schema in json string
	documentLoader := gojsonschema.NewStringLoader(schema)
	// This "schema" in schemaLoader is json schema not the compact schema
	schemaLoader := gojsonschema.NewStringLoader(validationSchema)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return false, nil, err
	}
	if isValid = result.Valid(); !isValid {
		for _, e := range result.Errors() {
			schemaErrors = append(schemaErrors, e.String())
		}
	}
	return isValid, schemaErrors, nil
}

func isBuiltInType(t string) bool {
	t = strings.ToLower(t)
	for _, bt := range builtinTypes {
		if t == strings.ToLower(bt) {
			return true
		}
	}
	return false
}

func isCompactName(namespace, typ string, compactNames map[string]Class) bool {
	for fullName := range compactNames {
		if typ == fullName {
			return true
		} else if !strings.Contains(typ, ".") {
			// If typ does not contain a dot, i.e it is not a full classname,
			// we also check if it is defined in the namespace of the schema that is being validated
			// This allows users to use short class names in their schema if they defined it in the schema.
			if namespace+"."+typ == fullName {
				return true
			}
		}
	}
	return false
}
