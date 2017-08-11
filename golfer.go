package golfer

import "github.com/gopherjs/gopherjs/js"

var document *js.Object
var body *js.Object

func init() {
	document = js.Global.Get("document")
	body = document.Get("body")
}

// Lib appends a JavaScript library to the DOM and loads it.
func Lib(src string) <-chan bool {
	loaded := make(chan bool, 1)
	script := document.Call("createElement", "script")
	script.Set("src", src)
	script.Set("onload", func() { loaded <- true })
	body.Call("appendChild", script)
	return loaded
}

// Text creates a text element.
func Text(name string) <-chan bool {
	text := document.Call("createElement", "text")
	text.Get("style").Set("visibility", "hidden")
	text.Set("name", name)
	clicked := make(chan bool, 1)
	text.Set("onclick", func() {
		text.Call("focus")
		clicked <- true
	})
	body.Call("appendChild", text)
	return clicked
}

// Callback returns a function that when run it fills the following channel.
func Callback() (func(), <-chan bool) {
	bc := make(chan bool, 1)
	return func() { bc <- true }, bc
}
