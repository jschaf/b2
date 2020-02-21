package main

import (
	"log"

	"github.com/jschaf/b2/pkg/markdown"
	"github.com/jschaf/b2/pkg/markdown/compiler"
)

func main() {
	md := markdown.New()
	c := compiler.New(md)
	if err := compiler.CompileEverything(c); err != nil {
		log.Fatal(err)
	}
}
