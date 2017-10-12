package jsutil

import (
	"errors"
	"strings"

	"github.com/gopherjs/gopherjs/js"
)

var (
	document  *js.Object
	body      *js.Object
	keyboard  *js.Object
	keybuffer chan string
)

func init() {
	defer func() {
		recover()
		return
	}()

	document = js.Global.Get("document")
	keyboard = document.Call("createElement", "input")
	keyboard.Set("type", "text")
	keyboard.Set("id", "keyboard")
	keyboard.Get("style").Set("top", -23)
	keyboard.Get("style").Set("opacity", 0.0)
	keyboard.Get("style").Set("position", "absolute")
	keyboard.Set("onclick", func() { keyboard.Call("focus") })
	body = document.Get("body")
	body.Call("appendChild", keyboard)
}

func extension(file string) string {
	return file[strings.LastIndex(file, "."):]
}

// Load appends a JavaScript library to the DOM and loads it.
func Load(url string, alt ...string) <-chan bool {
	loaded := make(chan bool)

	var file *js.Object

	switch extension(url) {
	case ".js":
		file = document.Call("createElement", "script")
		file.Set("src", url)
	case ".css":
		file = document.Call("createElement", "link")
		file.Set("rel", "stylesheet")
		file.Set("href", url)
	default:
		panic("bad file type")
	}

	file.Set("onload", F(func() { loaded <- true }))
	file.Set("onerror", F(func() {
		if len(alt) < 1 {
			loaded <- false
		} else {
			loaded <- <-Load(alt[0], alt[1:]...)
		}
	}))
	body.Call("appendChild", file)
	return loaded
}

// OpenKeyboard pulls up the soft keyboard.
func OpenKeyboard() <-chan string {
	keybuffer = make(chan string, 4)
	keyboard.Set("oninput", F(func() { keybuffer <- keyboard.Get("value").String() }))
	// keyboard.Set("onkeypress", fn)
	// keyboard.Set("onkeyup", fn)
	// keyboard.Set("onselect", fn)
	keyboard.Call("click")
	return keybuffer
}

// CloseKeyboard forces the soft keyboard away.
func CloseKeyboard() {
	close(keybuffer)
	keyboard.Set("oninput", nil)
	// keyboard.Set("onkeypress", nil)
	// keyboard.Set("onkeyup", nil)
	// keyboard.Set("onselect", nil)
	keyboard.Call("blur")
}

// F relieves callbacks completely from blocking.
func F(f func()) func() {
	return func() { go f() }
}

// C returns a function that when run it fills the following channel.
func C() (func(), <-chan bool) {
	c := make(chan bool)
	return F(func() { c <- true }), c
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
func OnPanic(err ...*error) {
	r := recover()

	if len(err) > 1 {
		Alert("OnPanic: too many arguments")
		return
	}

	if r == nil {
		*err[0] = nil
		return
	}

	if len(err) > 0 {
		switch e := r.(type) {
		case *js.Error:
			*err[0] = e
		case string:
			*err[0] = errors.New(e)
		case error:
			*err[0] = e
		default:
			Alert("OnPanic: unknown panic")
		}
		return
	}

	switch e := r.(type) {
	case *js.Error:
		Alert("panic: " + e.Error())
	case string:
		Alert("panic: " + e)
	case error:
		Alert("panic: " + e.Error())
	default:
		Alert("OnPanic: unknown panic")
	}
}
