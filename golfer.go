package golfer

import "github.com/gopherjs/gopherjs/js"

var document *js.Object
var body *js.Object
var keyboard *js.Object

func init() {
	document = js.Global.Get("document")
	body = document.Get("body")

	keyboard = document.Call("createElement", "text")
	keyboard.Get("style").Set("opacity", 0.0)
	keyboard.Get("style").Set("position", "absolute")
	keyboard.Set("id", "keyboard")
	keyboard.Set("onclick", func() {
		keyboard.Call("focus")
	})
	body.Call("appendChild", keyboard)
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

// OpenKeyboard pulls up the soft keyboard.
func OpenKeyboard() {
	keyboard.Call("click")
}

// ReadKeyboard reads the text buffer from key input after an open.
func ReadKeyboard() string {
	text := keyboard.Get("value").String()
	keyboard.Set("value", "")
	return text
}

// <input type="text" id="FirstName" onclick="focus()" style="opacity: 0.0; position:absolute">

// Callback returns a function that when run it fills the following channel.
func Callback() (func(), <-chan bool) {
	bc := make(chan bool, 1)
	return func() { bc <- true }, bc
}
