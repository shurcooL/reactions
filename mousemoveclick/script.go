// When mousemove handler calls setInnerHTML, it causes click handler not
// to trigger on iOS.
package main

import (
	"github.com/gopherjs/gopherjs/js"
	"honnef.co/go/js/dom"
)

var document = dom.GetWindow().Document().(dom.HTMLDocument)

func main() {
	document.AddEventListener("DOMContentLoaded", false, func(_ dom.Event) {
		setup()
	})
}

func setup() {
	label := document.CreateElement("div").(*dom.HTMLDivElement)

	box := document.CreateElement("div").(*dom.HTMLDivElement)
	box.Style().SetProperty("width", "400px", "")
	box.Style().SetProperty("height", "400px", "")
	box.Style().SetProperty("background-color", "red", "")

	box.AddEventListener("click", false, func(event dom.Event) {
		js.Global.Call("alert", "click")
	})
	box.AddEventListener("mousemove", false, func(event dom.Event) {
		js.Global.Call("alert", "mousemove")
		label.SetInnerHTML(`Some <b>stuff</b> here.`) // Doing this prevents click handler from triggerring on iOS/mobile.
	})

	document.Body().AppendChild(box)
	document.Body().AppendChild(label)
}
