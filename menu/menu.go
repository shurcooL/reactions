// +build js

package main

import (
	"fmt"
	"strings"

	"github.com/gopherjs/gopherjs/js"
	"github.com/shurcooL/reactions"
	"honnef.co/go/js/dom"
)

var document = dom.GetWindow().Document().(dom.HTMLDocument)

var Reactions ReactionsMenu

func main() {
	document.AddEventListener("DOMContentLoaded", false, func(dom.Event) {
		Reactions.authenticatedUser = true
		setupReactionsMenu()
		Reactions.Show(document.Body(), 0)
	})
}

func (rm ReactionsMenu) Show(this dom.HTMLElement, commentID uint64) {
	updateSelected(0)
	Reactions.filter.Value = ""
	Reactions.filter.Underlying().Call("dispatchEvent", js.Global.Get("CustomEvent").New("input")) // Trigger "input" event listeners.

	Reactions.menu.Style().SetProperty("display", "initial", "")

	// Reactions.menu aims to have 270px client width. Due to optional scrollbars
	// taking up some of that space, we may need to compensate and increase width.
	if scrollbarWidth := Reactions.results.OffsetWidth() - Reactions.results.Get("clientWidth").Float(); scrollbarWidth > 0 {
		Reactions.menu.Style().SetProperty("width", fmt.Sprintf("%fpx", 270+scrollbarWidth+1), "")
	}

	top := this.GetBoundingClientRect().Top - Reactions.menu.GetBoundingClientRect().Height - 8
	if top < 10 {
		top = 10
	}
	Reactions.menu.Style().SetProperty("top", fmt.Sprint(top), "")
	if rm.authenticatedUser {
		Reactions.filter.Focus()
	}
}

type ReactionsMenu struct {
	menu    *dom.HTMLDivElement
	filter  *dom.HTMLInputElement
	results *dom.HTMLDivElement

	authenticatedUser bool
}

func setupReactionsMenu() {
	Reactions.menu = document.CreateElement("div").(*dom.HTMLDivElement)
	Reactions.menu.SetID("rm-reactions-menu")

	container := document.CreateElement("div").(*dom.HTMLDivElement)
	container.SetClass("rm-reactions-menu-container")
	Reactions.menu.AppendChild(container)

	// Disable for unauthenticated user.
	if !Reactions.authenticatedUser {
		disabled := document.CreateElement("div").(*dom.HTMLDivElement)
		disabled.SetClass("rm-reactions-menu-disabled")
		signIn := document.CreateElement("div").(*dom.HTMLDivElement)
		signIn.SetClass("rm-reactions-menu-signin")
		signIn.SetInnerHTML(`<form method="post" action="/login/github" style="display: inline-block;"><input type="submit" name="" value="Sign in via GitHub"></form> to react.`)
		disabled.AppendChild(signIn)
		container.AppendChild(disabled)
	}

	Reactions.filter = document.CreateElement("input").(*dom.HTMLInputElement)
	Reactions.filter.SetClass("rm-reactions-filter")
	Reactions.filter.Placeholder = "Search"
	container.AppendChild(Reactions.filter)
	Reactions.results = document.CreateElement("div").(*dom.HTMLDivElement)
	Reactions.results.SetClass("rm-reactions-results")
	Reactions.results.AddEventListener("click", false, func(event dom.Event) {
		me := event.(*dom.MouseEvent)
		x := (me.ClientX - int(Reactions.results.GetBoundingClientRect().Left) + Reactions.results.Underlying().Get("scrollLeft").Int()) / 30
		if x >= 9 {
			return // Out of bounds to the right, likely because of scrollbar.
		}
		y := (me.ClientY - int(Reactions.results.GetBoundingClientRect().Top) + Reactions.results.Underlying().Get("scrollTop").Int()) / 30
		i := y*9 + x
		if i < 0 || i >= len(filtered) {
			return
		}
		emojiID := filtered[i]
		fmt.Printf("clicked %q reaction\n", emojiID)
	})
	container.AppendChild(Reactions.results)
	preview := document.CreateElement("div").(*dom.HTMLDivElement)
	container.AppendChild(preview)
	preview.SetOuterHTML(`<div class="rm-reactions-preview"><span id="rm-reactions-preview-emoji"></span><span id="rm-reactions-preview-label"></span></div>`)

	updateFilteredResults(Reactions.filter, Reactions.results)
	Reactions.filter.AddEventListener("input", false, func(dom.Event) {
		updateFilteredResults(Reactions.filter, Reactions.results)
	})

	Reactions.results.AddEventListener("mousemove", false, func(event dom.Event) {
		me := event.(*dom.MouseEvent)
		x := (me.ClientX - int(Reactions.results.GetBoundingClientRect().Left) + Reactions.results.Underlying().Get("scrollLeft").Int()) / 30
		if x >= 9 {
			return // Out of bounds to the right, likely because of scrollbar.
		}
		y := (me.ClientY - int(Reactions.results.GetBoundingClientRect().Top) + Reactions.results.Underlying().Get("scrollTop").Int()) / 30
		i := y*9 + x
		updateSelected(i)
	})

	document.Body().AppendChild(Reactions.menu)
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
		element.SetOuterHTML(`<div class="rm-reaction"><span class="rm-emoji" style="background-position: ` + reactions.Position(emojiID) + `;"></span></div>`)
		filtered = append(filtered, emojiID)
	}
}

// updateSelected reaction to filtered[index].
func updateSelected(index int) {
	if index < 0 || index >= len(filtered) {
		return
	}
	emojiID := filtered[index]

	label := document.GetElementByID("rm-reactions-preview-label").(*dom.HTMLSpanElement)
	label.SetTextContent(strings.Trim(emojiID, ":"))
	emoji := document.GetElementByID("rm-reactions-preview-emoji").(*dom.HTMLSpanElement)
	emoji.SetInnerHTML(`<span class="rm-emoji rm-large" style="background-position: ` + reactions.Position(emojiID) + `;"></span></div>`)
}
