package main

import (
	"log"
	"os"

	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	// Create a log file for debugging
	f, err := os.Create("/tmp/protoc-gen-graphql.log")
	if err == nil {
		defer f.Close()
		log.SetOutput(f)
		log.Println("Starting protoc-gen-graphql...")
	}

	protogen.Options{}.Run(func(gen *protogen.Plugin) error {
		if err != nil {
			log.Println("Error creating log file:", err)
		}
		log.Println("Plugin started")

		generator := newGenerator("/Users/fraser/code/federated-gql/tools/protoc-gen-graphql/templates/graphql-service-schema.tmpl")
		err := generator.Generate(gen)

		if err != nil {
			log.Println("Error generating:", err)
		} else {
			log.Println("Generation completed successfully")
		}

		return err
	})
}
