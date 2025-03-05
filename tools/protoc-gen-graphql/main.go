package main

import (
	"flag"
	"log"
	"os"

	"google.golang.org/protobuf/compiler/protogen"
)

// Options for the generator
type Options struct {
	TemplatePath string // Path to custom template file (falls back to embedded template if not provided)
}

func main() {
	log.SetPrefix("protoc-gen-graphql: ")
	log.SetFlags(0)
	log.SetOutput(os.Stderr)
	log.Println("Starting protoc-gen-graphql...")
	var flags flag.FlagSet

	opts := Options{}
	flags.StringVar(&opts.TemplatePath, "template_path", "", "Path to custom template file (falls back to embedded template if not provided)")

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(gen *protogen.Plugin) error {
		return newGenerator(opts).Generate(gen)
	})
}
