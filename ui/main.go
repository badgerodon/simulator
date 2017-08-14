package main

import (
	"fmt"
	"os/exec"

	"github.com/gopherjs/gopherjs/js"
)

//go:generate /Users/caleb/badgerodon/bin/gopherjs build -o assets/js/main.js github.com/badgerodon/simulator/ui
//go:generate sh -c "cp node_modules/xterm/dist/xterm.css assets/css/"
//go:generate sh -c "cp node_modules/xterm/dist/xterm.j* assets/js/"
//go:generate sh -c "cp node_modules/xterm/dist/addons/fit/fit.js assets/js/xterm.fit.js"
//go:generate echo done

func main() {
	//kernel.StartProcess("github.com/badgerodon/simulator-examples/hello", nil)
	cmd := exec.Command("github.com/badgerodon/simulator-examples/hello")
	cmd.Run()

	// root, _ := GetElementByID("root")
	// root.ReplaceWith(
	// 	E("div#root",
	// 		E("header",
	// 			E("h1", T("Simulator")),
	// 		),

	// 		E("section.section",
	// 			E("div.container",
	// 				E("div.columns",
	// 					E("div.column",
	// 						E("div#code"),
	// 					),
	// 				),
	// 			),
	// 		),

	// 		E("footer",
	// 			E("span", H("Copyright &copy;"+fmt.Sprint(time.Now().Year())+" Caleb Doxsey")),
	// 		),
	// 	),
	// )

	container := js.Global.Get("document").Call("getElementById", "terminal-container")

	term := js.Global.Get("Terminal").New()
	term.Call("open", container, false)
	for i := 0; i < 1000; i++ {
		term.Call("writeln", fmt.Sprint("Hello from \033[1;3;31mxterm.js\033[0m $ ", i))
	}

	onResize := func() {
		fontWidth, fontHeight := 10, 13
		ow, oh := container.Get("offsetWidth").Int(), container.Get("offsetHeight").Int()
		cols := ow / fontWidth
		rows := oh / fontHeight
		term.Call("resize", cols, rows)
	}
	onResize()
	js.Global.Set("onresize", onResize)

	js.Global.Get("console").Call("log", term)
}
