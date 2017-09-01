package jsutil

import "github.com/gopherjs/gopherjs/js"

var (
	document *js.Object
	body     *js.Object
	keyboard *js.Object
)

func init() {
	document = js.Global.Get("document")
	body = document.Get("body")

	keyboard = document.Call("createElement", "input")
	keyboard.Set("type", "text")
	keyboard.Set("id", "keyboard")
	keyboard.Get("style").Set("top", -23)
	keyboard.Get("style").Set("opacity", 0.0)
	keyboard.Get("style").Set("position", "absolute")
	keyboard.Set("onclick", func() { keyboard.Call("focus") })
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
func OpenKeyboard() <-chan string {
	txt := make(chan string, 4)
	keyboard.Set("oninput", func() { txt <- keyboard.Get("value").String() })
	// keyboard.Set("onkeypress", fn)
	// keyboard.Set("onkeyup", fn)
	// keyboard.Set("onselect", fn)
	keyboard.Call("click")
	return txt
}

// CloseKeyboard forces the soft keyboard away.
func CloseKeyboard() {
	keyboard.Set("oninput", nil)
	// keyboard.Set("onkeypress", nil)
	// keyboard.Set("onkeyup", nil)
	// keyboard.Set("onselect", nil)
	keyboard.Call("blur")
}

// Callback returns a function that when run it fills the following channel.
func Callback() (func(), <-chan bool) {
	bc := make(chan bool, 1)
	return func() { bc <- true }, bc
}
