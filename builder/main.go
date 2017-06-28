package main

import (
	"fmt"
	"html/template"
	"log"
	"os"
	"path/filepath"
	"time"
)

var working string

func main() {
	if os.Getenv("WORKING_DIR") != "" {
		working = os.Getenv("WORKING_DIR")
	} else {
		working = filepath.Join(os.TempDir(), fmt.Sprint(time.Now().UnixNano()))
		defer os.RemoveAll(working)
	}

	err := writeFiles()
	if err != nil {
		log.Fatalln(err)
	}
}

func writeFiles() error {
	err := os.MkdirAll(working, 0777)
	if err != nil {
		return err
	}

	for name, tpl := range map[string]*template.Template{
		"main.go":    mainTemplate,
		"main_js.go": mainJSTemplate,
		"env_js.go":  envJSTemplate,
	} {
		f, err := os.Create(filepath.Join(working, name))
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
