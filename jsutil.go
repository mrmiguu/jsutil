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
	keyboard.Get("style").Set("width", 0)
	keyboard.Get("style").Set("left", -10)
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
	input := make(chan string, 1)
	keyboard.Set("oninput", func() { input <- keyboard.Get("value").String() })
	keyboard.Call("click")
	return input
}

// Callback returns a function that when run it fills the following channel.
func Callback() (func(), <-chan bool) {
	bc := make(chan bool, 1)
	return func() { bc <- true }, bc
}

// OnPanic recovers on a panic, filling err.
func OnPanic(err *error) {
	e := recover()

	if e == nil {
		return
	}

	if e, ok := e.(*js.Error); ok {
		*err = e
	} else {
		panic(e)
	}
}
