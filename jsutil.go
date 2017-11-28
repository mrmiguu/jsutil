package jsutil

import (
	"compress/gzip"
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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

// FetchBlob reads a blob into a byte slice.
func FetchBlob(blob *js.Object) <-chan []byte {
	r := js.Global.Get("FileReader").New()
	c := make(chan []byte)
	r.Set("onload", F(func() {
		c <- js.Global.Get("Uint8Array").New(r.Get("result")).Interface().([]byte)
	}))
	r.Call("readAsArrayBuffer", blob)
	return c
}

// Compile compiles and minifies the path using GopherJS.
func Compile(path string) error {
	home, err := filepath.Abs(".")
	if err != nil {
		return err
	}
	dir, file := filepath.Split(path)

	// switch to the directory
	err = os.Chdir(dir)
	if err != nil {
		return err
	}

	// actually compile
	out, err := exec.Command("gopherjs", "build", "-m", file).Output()
	if err != nil {
		return err
	}
	if len(out) > 0 {
		return errors.New(string(out))
	}

	// switch back to home
	return os.Chdir(home)
}

// CompileWithGzip compiles, minifies, and compresses the path using GopherJS and gzip.
func CompileWithGzip(path string) error {
	err := Compile(path)
	if err != nil {
		return err
	}
	base := path[:strings.Index(path, filepath.Ext(path))]
	js, err := ioutil.ReadFile(base + ".js")
	if err != nil {
		return err
	}
	if err = os.Remove(base + ".js"); err != nil {
		return err
	}
	gz, err := os.Create(base + ".js.gz")
	if err != nil {
		return err
	}
	w := gzip.NewWriter(gz)
	if _, err = w.Write(js); err != nil {
		return err
	}
	if err = w.Close(); err != nil {
		return err
	}
	return gz.Close()
}

// FocusKeyboard pulls up the soft keyboard.
func FocusKeyboard() <-chan string {
	keyboard.Call("click")
	js.Global.Call("setTimeout", func() { keyboard.Call("click") }, 1) // Firefox is the only one that needs this
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
