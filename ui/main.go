package main

import (
	"log"
	"os/exec"

	_ "github.com/badgerodon/simulator/kernel"
	"github.com/gopherjs/gopherjs/js"
)

//go:generate gopherjs build -o assets/js/main.js github.com/badgerodon/simulator/ui
//go:generate sh -c "cp node_modules/xterm/dist/xterm.css assets/css/"
//go:generate sh -c "cp node_modules/xterm/dist/xterm.j* assets/js/"
//go:generate sh -c "cp node_modules/xterm/dist/addons/fit/fit.js assets/js/xterm.fit.js"
//go:generate echo done

type terminalWriter struct {
	term *js.Object
}

func (w *terminalWriter) Write(p []byte) (int, error) {
	//js.Global.Get("console").Call("log", "TW", "Write", p)
	w.term.Call("writeln", string(p))
	return len(p), nil
}

func main() {
	log.SetFlags(0)

	container := js.Global.Get("document").Call("getElementById", "terminal-container")

	term := js.Global.Get("Terminal").New()
	term.Call("open", container, false)
	onResize := func() {
		fontWidth, fontHeight := 10, 13
		ow, oh := container.Get("offsetWidth").Int(), container.Get("offsetHeight").Int()
		cols := ow / fontWidth
		rows := oh / fontHeight
		term.Call("resize", cols, rows)
	}
	onResize()
	js.Global.Set("onresize", onResize)

	w := &terminalWriter{
		term: term,
	}

	//kernel.StartProcess("github.com/badgerodon/simulator-examples/hello", nil)
	cmd := exec.Command(js.Global.Get("location").Get("pathname").Call("substr", 5).String())
	cmd.Stdout = w
	cmd.Stderr = w
	err := cmd.Run()

	if err != nil {
		term.Call("writeln", err.Error())
	}
}
