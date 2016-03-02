// +build js

package main

import (
	"strings"

	"github.com/shurcooL/reactions"
	"honnef.co/go/js/dom"
)

var document = dom.GetWindow().Document().(dom.HTMLDocument)

func main() {
	document.AddEventListener("DOMContentLoaded", false, func(dom.Event) {
		setup()
	})
}

func setup() {
	overlay := document.CreateElement("div").(*dom.HTMLDivElement)
	overlay.SetClass("reactions-menu")

	container := document.CreateElement("div")
	overlay.AppendChild(container)

	filter := document.CreateElement("input").(*dom.HTMLInputElement)
	filter.SetClass("reactions-filter")
	filter.Placeholder = "Search"
	container.AppendChild(filter)
	results := document.CreateElement("div").(*dom.HTMLDivElement)
	results.SetClass("reactions-results")
	container.AppendChild(results)

	update(filter, results)
	filter.AddEventListener("input", false, func(dom.Event) {
		update(filter, results)
	})

	document.Body().AppendChild(overlay)
}

func update(filter *dom.HTMLInputElement, results dom.Element) {
	lower := strings.ToLower(strings.TrimSpace(filter.Value))
	results.SetInnerHTML("")
	for _, emojiID := range reactions.Sorted {
		if lower != "" && !strings.Contains(emojiID, lower) {
			continue
		}
		element := document.CreateElement("div")
		results.AppendChild(element)
		element.SetOuterHTML(`<div class="reaction"><span class="emoji" style="background-position: ` + reactions.Position(emojiID) + `;"></span></div>`)
	}
}
