package serializer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

func generateCompactClasses(lang string, schema Schema) (map[ClassInfo]string, error) {
	switch lang {
	case langJava:
		return generateJavaClasses(schema)
	default:
		return nil, fmt.Errorf("unsupported langugage: %s", lang)
	}
}

func saveCompactClasses(outputDir string, classes map[ClassInfo]string) error {
	err := os.MkdirAll(outputDir, fs.ModePerm)
	if err != nil {
		return fmt.Errorf("generating target directories at path %s: %w", outputDir, err)
	}
	for k, v := range classes {
		p := filepath.Join(outputDir, k.FileName)
		err = os.WriteFile(p, []byte(v), fs.ModePerm)
		if err != nil {
			return err
		}
	}
	return nil
}
