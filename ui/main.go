package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
)

//go:generate gopherjs build -o dist/bundle.js .
//go:generate sh -c "cat dist/bundle.js | head -n -1 > dist/tmp.js"
//go:generate sh -c "mv dist/tmp.js dist/bundle.js"
//go:generate rm dist/bundle.js.map

type root struct {
	vecty.Core
}

func (r *root) Render() *vecty.HTML {
	return elem.Body(
		elem.Heading1(
			vecty.Text("Simulator"),
		),
		elem.Div(vecty.Attribute("id", "container"), vecty.Style("width", "800px"), vecty.Style("height", "800px")),
		elem.Script(vecty.Text(`
			require(["vs/editor/editor.main"], function () {
				var editor = monaco.editor.create(document.getElementById('container'), {
					value: [
						'function x() {',
						'\tconsole.log("Hello world!");',
						'}'
					].join('\n'),
					language: 'javascript'
				});
			});
		`)),
	)
}

func main() {
	//kernel.StartProcess("github.com/badgerodon/simulator-examples/hello", nil)

	js.Global.Get("requirejs").Invoke([]string{"github-api"}, func(GitHub *js.Object) {
		gh := GitHub.New(js.M{})

		repo := gh.Call("getRepo", "badgerodon", "simulator-examples")
		repo.Call("getContents", "master", "hello/main.go", true, func(err, res, req *js.Object) {
			js.Global.Get("console").Call("log", err, res)
		})
	})

	vecty.SetTitle("Simulator")
	vecty.RenderBody(new(root))
}
