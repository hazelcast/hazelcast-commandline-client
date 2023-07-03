package project

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func applyTemplateAndCopyToTarget(vars map[string]string, source, dest string) error {
	destFile, err := os.Create(removeFileExt(dest))
	if err != nil {
		log.Fatal(err)
	}
	defer destFile.Close()
	tmpl, err := template.ParseFiles(source)
	if err != nil {
		return err
	}
	err = tmpl.Execute(destFile, vars)
	if err != nil {
		return err
	}
	return nil
}

func copyToTarget(source string, dest string, removeExt bool) error {
	sourceFile, err := os.Open(source)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	if removeExt {
		dest = removeFileExt(dest)
	}
	destinationFile, err := os.Create(dest)
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

func removeFileExt(dest string) string {
	return strings.TrimSuffix(dest, filepath.Ext(dest))
}
