package generate

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"text/template"

	"github.com/hazelcast/hazelcast-commandline-client/internal/plug"
)

func templateExists(tDir, t string) (bool, error) {
	if _, err := os.Stat(filepath.Join(tDir, t)); err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func applyTemplateAndCopyToTarget(ec plug.ExecContext, tDir, fileName, destDir string) error {
	destFile, err := os.Create(filepath.Join(destDir, fileName))
	if err != nil {
		log.Fatal(err)
	}
	defer destFile.Close()
	tmpl, err := template.ParseFiles(filepath.Join(tDir, fileName))
	if err != nil {
		return err
	}
	data := make(map[string]string)
	err = loadFromDefaultProperties(tDir, &data)
	if err != nil {
		return err
	}
	loadFromProps(ec, &data)
	loadFromUserInput(ec, &data)
	err = tmpl.Execute(destFile, data)
	if err != nil {
		return err
	}
	return nil
}

func copyToTarget(tDir, fileName, destinationDir string) error {
	sourceFile, err := os.Open(filepath.Join(tDir, fileName))
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	destinationFile, err := os.Create(filepath.Join(destinationDir, fileName))
	if err != nil {
		return err
	}
	defer destinationFile.Close()
	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return err
	}
	return nil
}
