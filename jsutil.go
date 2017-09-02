package jsutil

import "github.com/gopherjs/gopherjs/js"

var (
	document *js.Object
	body     *js.Object
	keyboard *js.Object
)

func init() {
	defer func() {
		recover()
		return
	}()

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

// Alert calls the global alert function
func Alert(s string) {
	js.Global.Call("alert", s)
}

// Prompt prompts the user for a response with an optional title.
func Prompt(s ...string) string {
	msg := ""
	if len(s) > 0 {
		msg = s[0]
	}
	return js.Global.Call("prompt", msg).String()
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
