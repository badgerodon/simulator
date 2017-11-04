package builder

import (
	"html/template"
	"os"
	"path/filepath"
)

var (
	initTemplate = template.Must(template.New("").Parse(`package main

import (
    _ "github.com/badgerodon/simulator/kernel"
)

`))
)

func writeFiles(projectDir string) error {
	for name, tpl := range map[string]*template.Template{
		"aaa_simulator_kernel_init.go": initTemplate,
	} {
		f, err := os.Create(filepath.Join(projectDir, name))
		if err != nil {
			return err
		}
		defer f.Close()
		err = tpl.Execute(f, nil)
		if err != nil {
			return err
		}
	}

	return nil
}
