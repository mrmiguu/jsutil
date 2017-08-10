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

// Callback returns a function that when run it fills the following channel.
func Callback() (func(), <-chan bool) {
	bc := make(chan bool)
	return func() { bc <- true }, bc
}
