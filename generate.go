// +build generate

package main

import (
	"log"

	"github.com/google/go-github/github"
	"github.com/shurcooL/go-goon"
)

func run() error {
	gh := github.NewClient(nil)
	emojis, _, err := gh.ListEmojis()
	if err != nil {
		return err
	}

	goon.DumpExpr(emojis)

	return nil
}

func main() {
	err := run()
	if err != nil {
		log.Fatalln(err)
	}
}
