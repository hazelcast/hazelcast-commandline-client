package serializer

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"strings"
)

func generateCompactClasses(lang string, schema Schema) (map[ClassInfo]string, error) {
	classes := make(map[ClassInfo]string)
	switch lang {
	case java:
		generateJavaClasses(schema, classes)
	case python:
		panic(any("implement me"))
	case typescript:
		panic(any("implement me"))
	case cpp:
		panic(any("implement me"))
	case golang:
		panic(any("implement me"))
	case cs:
		panic(any("implement me"))
	default:
		return nil, fmt.Errorf("unsupported langugage")
	}
	return classes, nil
}

func saveCompactClasses(outputDir string, classes map[ClassInfo]string) error {
	err := os.MkdirAll(outputDir, fs.ModePerm)
	if err != nil {
		return fmt.Errorf("generating target directories at path %s: %w", outputDir, err)
	}
	var errString strings.Builder
	for k, v := range classes {
		p := path.Join(outputDir, k.FileName)
		err := os.WriteFile(p, []byte(v), fs.ModePerm)
		if err != nil {
			errString.WriteString(err.Error() + "\n")
		}
	}
	if errString.String() != "" {
		return fmt.Errorf(errString.String())
	}
	return nil
}
