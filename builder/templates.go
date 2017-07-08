package builder

import (
	"html/template"
	"os"
	"path/filepath"
)

var (
	initTemplate = template.Must(template.New("").Parse(`package main

import (
    "fmt"
    "github.com/gopherjs/gopherjs/js"
	"net/url"
    "strings"
)

func main() {
    u, _ := url.Parse(js.Global.Get("location").Get("href").String())
    for k, vs := range u.Query() {
        os.Setenv(k, strings.Join(vs, ","))
    }
    fmt.Println(os.Environ())
}
`))
)

func writeFiles(projectDir string) error {
	err := os.MkdirAll(projectDir, 0777)
	if err != nil {
		return err
	}

	for name, tpl := range map[string]*template.Template{
		"simulator_init.go": initTemplate,
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
