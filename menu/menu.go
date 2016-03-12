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

	container := document.CreateElement("div").(*dom.HTMLDivElement)
	container.SetClass("reactions-menu-container")
	overlay.AppendChild(container)

	const authenticatedUser = false
	// Disable for unauthenticated user.
	if !authenticatedUser {
		disabled := document.CreateElement("div").(*dom.HTMLDivElement)
		disabled.SetClass("reactions-menu-disabled")
		signIn := document.CreateElement("div").(*dom.HTMLDivElement)
		signIn.SetClass("reactions-menu-signin")
		signIn.SetInnerHTML(`<form method="post" action="/login/github" style="display: inline-block;"><input type="submit" name="" value="Sign in via GitHub"></form> to react.`)
		disabled.AppendChild(signIn)
		container.AppendChild(disabled)
	}

	filter := document.CreateElement("input").(*dom.HTMLInputElement)
	filter.SetClass("reactions-filter")
	filter.Placeholder = "Search"
	container.AppendChild(filter)
	results := document.CreateElement("div").(*dom.HTMLDivElement)
	results.SetClass("reactions-results")
	container.AppendChild(results)
	preview := document.CreateElement("div").(*dom.HTMLDivElement)
	container.AppendChild(preview)
	preview.SetOuterHTML(`<div class="reactions-preview"><span id="reactions-preview-emoji"></span><span id="reactions-preview-label"></span></div>`)

	updateFilteredResults(filter, results)
	filter.AddEventListener("input", false, func(dom.Event) {
		updateFilteredResults(filter, results)
	})

	results.AddEventListener("mousemove", false, func(event dom.Event) {
		me := event.(*dom.MouseEvent)
		x := (me.ClientX - int(results.GetBoundingClientRect().Left) + results.Underlying().Get("scrollLeft").Int()) / 30
		y := (me.ClientY - int(results.GetBoundingClientRect().Top) + results.Underlying().Get("scrollTop").Int()) / 30
		i := y*9 + x
		updateSelected(i)
	})

	document.Body().AppendChild(overlay)

	updateSelected(0)
}

var filtered []string

func updateFilteredResults(filter *dom.HTMLInputElement, results dom.Element) {
	lower := strings.ToLower(strings.TrimSpace(filter.Value))
	results.SetInnerHTML("")
	filtered = nil
	for _, emojiID := range reactions.Sorted {
		if lower != "" && !strings.Contains(emojiID, lower) {
			continue
		}
		element := document.CreateElement("div")
		results.AppendChild(element)
		element.SetOuterHTML(`<div class="reaction"><span class="emoji" style="background-position: ` + reactions.Position(emojiID) + `;"></span></div>`)
		filtered = append(filtered, emojiID)
	}
}

// updateSelected reaction to filtered[index].
func updateSelected(index int) {
	if index < 0 || index >= len(filtered) {
		return
	}
	emojiID := filtered[index]

	label := document.GetElementByID("reactions-preview-label").(*dom.HTMLSpanElement)
	label.SetTextContent(strings.Trim(emojiID, ":"))
	emoji := document.GetElementByID("reactions-preview-emoji").(*dom.HTMLSpanElement)
	emoji.SetInnerHTML(`<span class="emoji large" style="background-position: ` + reactions.Position(emojiID) + `;"></span></div>`)
}
