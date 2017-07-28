package main

import (
	"github.com/badgerodon/simulator/kernel"
	"github.com/gopherjs/vecty"
	"github.com/gopherjs/vecty/elem"
)

//go:generate gopherjs build -o dist/bundle.js .

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
