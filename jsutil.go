package jsutil

import (
	"errors"
	"strings"

	"github.com/gopherjs/gopherjs/js"
)

var (
	window    *js.Object
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

	window = js.Global.Get("window")
	document = js.Global.Get("document")
	body = document.Get("body")

	keyboard = document.Call("createElement", "input")
	keyboard.Set("type", "text")
	keyboard.Set("id", "keyboard")
	keyboard.Get("style").Set("top", -230)
	keyboard.Get("style").Set("opacity", 0.0)
	keyboard.Get("style").Set("position", "absolute")
	keyboard.Set("onclick", func() { keyboard.Call("focus") })
	keyboard.Set("onblur", func(e *js.Object) {
		e.Call("preventDefault")
	})
	keyboard.Set("oninput", func() { go func() { keybuffer <- keyboard.Get("value").String() }() })
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

// Redirect changes the browser's current web page.
func Redirect(url string) {
	window.Get("location").Set("href", url)
}

// Cookie is a map of cookie keys to values.
type Cookie map[string]string

// StoreCookie sets the current document's cookie to all key-values.
func StoreCookie(c Cookie) {
	var kvs []string
	for k, v := range c {
		kvs = append(kvs, k+"="+v)
	}
	document.Set("cookie", strings.Join(kvs, "%"))
}

// LoadCookie gets the current document's cookie of all key-values.
func LoadCookie() Cookie {
	c := Cookie{}
	cookie := document.Get("cookie").String()
	if len(cookie) == 0 {
		return c
	}
	println("cookie", cookie)
	kvs := strings.Split(cookie, "%")
	for _, kv := range kvs {
		piv := strings.Index(kv, "=")
		c[kv[:piv]] = kv[piv+1:]
	}
	return c
}

// ReadBlob reads a blob into a byte slice.
func ReadBlob(blob *js.Object) []byte {
	r := js.Global.Get("FileReader").New()
	f, c := C()
	r.Set("onload", f)
	r.Call("readAsArrayBuffer", blob)
	<-c
	return js.Global.Get("Uint8Array").New(r.Get("result")).Interface().([]byte)
}

// FocusKeyboard pulls up the soft keyboard.
func FocusKeyboard() <-chan string {
	keyboard.Call("click")
	js.Global.Call("setTimeout", func() { keyboard.Call("click") }, 1) // Firefox is the only one that needs this
	keyboard.Set("value", "")
	// keyboard.Set("onkeypress", fn)
	// keyboard.Set("onkeyup", fn)
	// keyboard.Set("onselect", fn)
	return keybuffer
}

// ClearKeyboard clears the soft keyboard buffer.
func ClearKeyboard() {
	keyboard.Set("value", "")
}

// BlurKeyboard forces the soft keyboard away.
func BlurKeyboard() {
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
func C() (func(), <-chan bool) {
	c := make(chan bool)
	return func() { go func() { c <- true }() }, c
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
