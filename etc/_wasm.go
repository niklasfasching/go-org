package main

import (
	"fmt"
	"strings"
	"syscall/js"

	"github.com/niklasfasching/go-org/org"
)

func main() {
	js.Global().Call("initialized")

	document := js.Global().Get("document")
	in := document.Call("getElementById", "input")
	out := document.Call("getElementById", "output")

	js.Global().Set("run", js.NewCallback(func([]js.Value) {
		in := strings.NewReader(in.Get("value").String())
		html, err := org.NewDocument().Parse(in).Write(org.NewHTMLWriter())
		if err != nil {
			out.Set("innerHTML", fmt.Sprintf("<pre>%s</pre>", err))
		} else {
			out.Set("innerHTML", html)
		}
	}))

	<-make(chan struct{}) // stay alive
}
