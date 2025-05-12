package main

import (
	"fmt"
	"strings"
	"syscall/js"

	"github.com/niklasfasching/go-org/org"
)

func main() {
	js.Global().Call("initialized")
	doc := js.Global().Get("document")
	in, out := doc.Call("getElementById", "input"), doc.Call("getElementById", "output")
	js.Global().Set("run", js.FuncOf(func(js.Value, []js.Value) interface{} {
		in := strings.NewReader(in.Get("value").String())
		html, err := org.New().Parse(in, "").Write(org.NewHTMLWriter())
		if err != nil {
			out.Set("innerHTML", fmt.Sprintf("<pre>%s</pre>", err))
		} else {
			out.Set("innerHTML", html)
		}
		return nil
	}))

	select {} // stay alive
}
