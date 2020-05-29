package main

import (
	"flag"
	"log"

	"github.com/jschaf/b2/pkg/logs"
	"github.com/jschaf/b2/pkg/markdown"
	"github.com/jschaf/b2/pkg/markdown/compiler"
	"github.com/jschaf/b2/pkg/markdown/mdext"
	"github.com/jschaf/b2/pkg/static"
)

var flagGlob = flag.String(
	"glob", "",
	"Only compile posts with an exact substring match on the filename")

func main() {
	flag.Parse()
	logger, err := logs.NewShortDevLogger()
	if err != nil {
		log.Fatalf("create dev logger: %s", err)
	}
	c := compiler.New(markdown.New(
		logger,
		markdown.WithExtender(mdext.NewNopContinueReadingExt())))
	if err := c.CompileAllPosts(*flagGlob); err != nil {
		log.Fatal(err)
	}

	ic := compiler.NewForIndex(
		markdown.New(logger, markdown.WithExtender(mdext.NewContinueReadingExt())))
	if err := ic.Compile(); err != nil {
		log.Fatal(err)
	}

	if err := static.CopyStaticFiles(); err != nil {
		log.Fatal(err)
	}

	if err := static.LinkPapers(); err != nil {
		log.Fatal(err)
	}
}
