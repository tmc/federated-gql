package main

import (
	"log"
	"os"
	"path/filepath"
	"runtime"

	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	// Set log output to stderr for better debugging visibility
	log.SetOutput(os.Stderr)
	log.Println("Starting protoc-gen-graphql...")

	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		log.Println("Plugin started")

		// Get the absolute path to the template
		_, filename, _, ok := runtime.Caller(0)
		if !ok {
			log.Println("Failed to get current file path")
			return nil
		}

		// Calculate the template path relative to the current file
		templatePath := filepath.Join(filepath.Dir(filename), "templates/graphql-service-schema.tmpl")
		log.Println("Using template path:", templatePath)

		generator := newGenerator(templatePath)
		err := generator.Generate(gen)

		if err != nil {
			log.Println("Error generating:", err)
		} else {
			log.Println("Generation completed successfully")
		}

		return err
	})
}
