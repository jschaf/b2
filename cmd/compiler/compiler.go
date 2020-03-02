package main

import (
	"flag"
	"log"

	"github.com/jschaf/b2/pkg/markdown"
	"github.com/jschaf/b2/pkg/markdown/compiler"
	"github.com/jschaf/b2/pkg/markdown/mdext"
)

var flagGlob = flag.String(
	"glob", "",
	"Only compile posts with an exact substring match on the filename")

func main() {
	flag.Parse()
	c := compiler.New(
		markdown.New(mdext.NewNopContinueReadingExt()))
	if err := c.CompileAllPosts(*flagGlob); err != nil {
		log.Fatal(err)
	}

	ic := compiler.NewForIndex(
		markdown.New(mdext.NewContinueReadingExt()))
	if err := ic.Compile(); err != nil {
		log.Fatal(err)
	}
}
