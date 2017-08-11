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
		elem.Div(vecty.Attribute("class", "container"),
			elem.Heading1(
				vecty.Text("Simulator"),
			),
		),
		elem.Section(
			elem.Div(vecty.Attribute("class", "container"),
				elem.Div(vecty.Attribute("class", "columns"),
					elem.Div(vecty.Attribute("class", "column"),
						elem.Preformatted(
							elem.Code(vecty.Attribute("id", "code"), vecty.Attribute("class", "lang-go")),
							vecty.Style("font-family", "Go Mono"),
							vecty.Style("font-size", "10px"),
						),
					),
					elem.Div(vecty.Attribute("class", "column")),
				),
			),
		),
	)
}

func main() {
	//kernel.StartProcess("github.com/badgerodon/simulator-examples/hello", nil)

	vecty.SetTitle("Simulator")
	vecty.RenderBody(new(root))
	js.Global.Get("requirejs").Invoke([]string{
		"github-api", "highlight",
	}, func(GitHub *js.Object, hljs *js.Object) {
		js.Global.Get("requirejs").Invoke([]string{
			"highlight-go",
		}, func() {
			gh := GitHub.New(js.M{})

			repo := gh.Call("getRepo", "badgerodon", "simulator-examples")
			repo.Call("getContents", "master", "hello/main.go", true, func(err, res, req *js.Object) {
				el := js.Global.Get("document").Call("getElementById", "code")
				el.Set("innerHTML", res)
				hljs.Call("highlightBlock", el)
			})
		})
	})
}
