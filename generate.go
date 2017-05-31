// +build generate

package main

import (
	"context"
	"log"

	"github.com/google/go-github/github"
	"github.com/shurcooL/go-goon"
)

func main() {
	err := run()
	if err != nil {
		log.Fatalln(err)
	}
}

func run() error {
	gh := github.NewClient(nil)
	emojis, _, err := gh.ListEmojis(context.Background())
	if err != nil {
		return err
	}

	goon.DumpExpr(emojis)

	return nil
}
