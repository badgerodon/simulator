package meshjs

import "github.com/gopherjs/gopherjs/js"

func log(args ...interface{}) {
	js.Global.Get("console").Call("log", args...)
}

func warn(args ...interface{}) {
	js.Global.Get("console").Call("warn", args...)
}
