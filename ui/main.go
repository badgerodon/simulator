package main

import (
	"github.com/badgerodon/simulator/kernel"
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
	)
}

func main() {
	kernel.StartProcess("/build/github.com/badgerodon/simulator-examples/hello", nil)

	vecty.SetTitle("Simulator")
	vecty.RenderBody(new(root))
}
