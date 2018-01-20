package main

import (
	"bufio"
	"bytes"
	"log"
	"os"
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
	s := bufio.NewScanner(bytes.NewReader(p))
	for s.Scan() {
		w.term.Call("writeln", s.Text())
	}
	return len(p), nil
}

func main() {
	log.SetFlags(0)

	container := js.Global.Get("document").Call("getElementById", "terminal-container")

	stdinr, stdinw, err := os.Pipe()
	if err != nil {
		log.Fatal(err)
	}
	defer stdinr.Close()
	defer stdinw.Close()

	term := js.Global.Get("Terminal").New(js.M{
		"cursorBlink": true,
	})
	term.Call("on", "data", func(data *js.Object) {
		go func() {
			stdinw.Write([]byte(data.String()))
		}()
	})
	term.Call("on", "key", func(key *js.Object, evt *js.Object) {
		printable := !evt.Get("altKey").Bool() &&
			!evt.Get("altGraphKey").Bool() &&
			!evt.Get("ctrlKey").Bool() &&
			!evt.Get("metaKey").Bool()

		//js.Global.Get("console").Call("log", "KEY", evt.Get("keyCode").Int())

		switch evt.Get("keyCode").Int() {
		case 8:
			term.Call("write", "\b \b")
		case 13:
			term.Call("writeln", "")
		default:
			if printable {
				term.Call("write", key)
			}
		}
	})
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
	cmd.Stdin = stdinr
	cmd.Env = append(cmd.Env, "TERM=dumb")
	err = cmd.Run()

	if err != nil {
		term.Call("writeln", err.Error())
	}
}
