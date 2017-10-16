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
	keybuffer = make(chan string, 4)
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
	keyboard.Get("style").Set("top", -230)
	keyboard.Get("style").Set("opacity", 0.0)
	keyboard.Get("style").Set("position", "absolute")
	keyboard.Set("onclick", func() { keyboard.Call("focus") })
	keyboard.Set("oninput", F(func() { keybuffer <- keyboard.Get("value").String() }))
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
	keyboard.Call("click")
	js.Global.Call("setTimeout", func() { keyboard.Call("click") }, 1) // Firefox is the only one that needs this
	// keyboard.Set("onkeypress", fn)
	// keyboard.Set("onkeyup", fn)
	// keyboard.Set("onselect", fn)
	return keybuffer
}

// CloseKeyboard forces the soft keyboard away.
func CloseKeyboard() {
	keyboard.Call("blur")
	// keyboard.Set("onkeypress", nil)
	// keyboard.Set("onkeyup", nil)
	// keyboard.Set("onselect", nil)
}

// F relieves callbacks completely from blocking.
func F(f func()) func() {
	return func() { go f() }
}

// C returns a function that when run it fills the following channel.
func C() (func(...*js.Object), <-chan []*js.Object) {
	c := make(chan []*js.Object)
	return func(args ...*js.Object) { go func() { c <- args }() }, c
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

// Panic panics in a JavaScript-friendly way.
func Panic(arg interface{}) {
	var err string
	switch x := arg.(type) {
	case string:
		err = x
	case *js.Error:
		err = x.Error()
	case error:
		err = x.Error()
	default:
		err = "unknown panic"
	}
	Alert("panic: " + err)
	panic(err)
}

// OnPanic recovers on a panic, filling err.
func OnPanic(err ...*error) {
	r := recover()

	if len(err) > 1 {
		Alert("OnPanic: too many arguments")
		panic("too many arguments")
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
			panic("unknown panic")
		}
		return
	}

	Panic(r)
}
