package main

import (
	"log"

	"github.com/jschaf/b2/pkg/markdown"
	"github.com/jschaf/b2/pkg/markdown/compiler"
	"github.com/jschaf/b2/pkg/markdown/mdext"
)

func main() {
	c := compiler.New(
		markdown.New(mdext.NewNopContinueReadingExt()))
	if err := c.CompileAllPosts(); err != nil {
		log.Fatal(err)
	}

	ic := compiler.NewForIndex(
		markdown.New(mdext.NewContinueReadingExt()))
	if err := ic.Compile(); err != nil {
		log.Fatal(err)
	}
}
