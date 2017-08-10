package golfer

import "github.com/gopherjs/gopherjs/js"

// Lib appends a JavaScript library to the DOM and loads it.
func Lib(src string) <-chan bool {
	loaded := make(chan bool)
	document := js.Global.Get("document")
	script := document.Call("createElement", "script")
	script.Set("src", src)
	script.Set("onload", func() { loaded <- true })
	document.Get("body").Call("appendChild", script)
	return loaded
}
