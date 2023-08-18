//go:build std || serializer

package serializer

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func validateInputs(schemaPath, language string) error {
	if !isLangSupported(language) {
		return fmt.Errorf("unsupported language, you can provide one of %s", strings.Join(supportedLanguages, ","))
	}
	if err := isSchemaExists(schemaPath); err != nil {
		return err
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

func isSchemaExists(schemaPath string) error {
	_, err := os.Stat(schemaPath)
	if errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("path %s does not exist", schemaPath)
	}
	if err != nil {
		return fmt.Errorf("can not access output directory %s, err: %s", schemaPath, err.Error())
	}
	return nil
}

func validateWithJSONSchema(schema map[string]any) (isValid bool, schemaErrors []string, err error) {
	return true, nil, nil
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
		}
		if !strings.Contains(typ, ".") {
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
