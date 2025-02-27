package main

import (
	"embed"
	"fmt"
	"log"
	"regexp"
	"strings"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

// Generator handles the generation of GraphQL schema files from proto definitions
type Generator struct {
	TemplateDir string
}

func newGenerator(templateDir string) *Generator {
	return &Generator{
		TemplateDir: templateDir,
	}
}

// Generate processes protobuf files and generates the corresponding GraphQL schema
func (g *Generator) Generate(gen *protogen.Plugin) error {
	gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
	
	log.Println("Starting generation...")
	log.Printf("Processing %d files\n", len(gen.Files))

	for i, f := range gen.Files {
		log.Printf("File %d: %s, Generate: %v\n", i, f.Desc.Path(), f.Generate)
		if !f.Generate {
			continue
		}

		log.Printf("File %s has %d services\n", f.Desc.Path(), len(f.Services))
		for j, svc := range f.Services {
			log.Printf("Processing service %d: %s\n", j, svc.Desc.FullName())
			if err := g.generateServiceSchema(svc, gen); err != nil {
				log.Printf("Error generating schema for service %s: %v\n", svc.Desc.FullName(), err)
				return err
			}
		}
	}

	log.Println("Generation completed successfully")
	return nil
}

//go:embed templates/*
var defaultTemplates embed.FS

// renderTemplate loads and executes a template with the provided data
func (g *Generator) renderTemplate(templateName string, service *protogen.Service, gf *protogen.GeneratedFile) error {
	log.Printf("Rendering template %s for service %s\n", templateName, service.Desc.FullName())
	
	// Read template content directly from embedded FS
	log.Printf("Reading template from templates/%s\n", templateName)
	content, err := defaultTemplates.ReadFile("templates/" + templateName)
	if err != nil {
		log.Printf("Error reading template: %v\n", err)
		return fmt.Errorf("failed to read template: %v", err)
	}
	log.Printf("Template content length: %d bytes\n", len(content))

	// Create and parse template
	t := template.New(templateName).Funcs(template.FuncMap{
		"lower":      strings.ToLower,
		"trim":       strings.TrimSpace,
		"trimPrefix": strings.TrimPrefix,
		"trimSuffix": strings.TrimSuffix,
		"hasPrefix":  strings.HasPrefix,
		"hasSuffix":  strings.HasSuffix,
		"replaceAll": strings.ReplaceAll,
		"pascal":     toPascalCase,
		"camel":      toCamelCase,
		"snake":      toSnakeCase,
		"getId":      getIdField,
		"getEntity":  getEntityMessage,
	})

	log.Println("Parsing template")
	t, err = t.Parse(string(content))
	if err != nil {
		log.Printf("Error parsing template: %v\n", err)
		return fmt.Errorf("failed to parse template: %v", err)
	}

	// Execute template
	log.Println("Executing template")
	err = t.Execute(gf, service)
	if err != nil {
		log.Printf("Error executing template: %v\n", err)
		return err
	}
	
	log.Println("Template rendered successfully")
	return nil
}

// generateServiceSchema creates a GraphQL schema for the protobuf service
func (g *Generator) generateServiceSchema(svc *protogen.Service, gen *protogen.Plugin) error {
	serviceName := string(svc.Desc.FullName())
	gf := gen.NewGeneratedFile(fmt.Sprintf("%s.graphql", serviceName), protogen.GoImportPath(serviceName))

	return g.renderTemplate("graphql-service-schema.tmpl", svc, gf)
}

// getServiceName extracts a simple service name from the full name
func getServiceName(svc *protogen.Service) string {
	fullName := string(svc.Desc.FullName())
	parts := strings.Split(fullName, ".")
	return parts[len(parts)-1]
}

// getIdField returns the ID field of a message
func getIdField(svc *protogen.Service) string {
	for _, m := range svc.Methods {
		if strings.HasPrefix(string(m.Desc.Name()), "Get") {
			for _, f := range m.Input.Fields {
				if strings.HasSuffix(string(f.Desc.Name()), "_id") {
					// Convert snake_case to camelCase and ensure "ID" is properly cased
					name := string(f.Desc.Name())
					name = toCamelCase(name)
					name = strings.ReplaceAll(name, "Id", "ID")
					return name
				}
			}
		}
	}
	return "id"
}

// getEntityMessage returns the entity message from a service
func getEntityMessage(svc *protogen.Service) *protogen.Message {
	// First look for a response type from a Get method
	for _, m := range svc.Methods {
		if strings.HasPrefix(string(m.Desc.Name()), "Get") {
			// Check the response message for an entity field
			for _, f := range m.Output.Fields {
				if f.Message != nil && !strings.HasSuffix(string(f.Message.Desc.Name()), "Request") &&
					!strings.HasSuffix(string(f.Message.Desc.Name()), "Response") {
					return f.Message
				}
			}
		}
	}
	
	// If not found, look for any message that's not a request/response
	for _, m := range svc.Methods {
		for _, f := range m.Output.Fields {
			if f.Message != nil && !strings.HasSuffix(string(f.Message.Desc.Name()), "Request") &&
				!strings.HasSuffix(string(f.Message.Desc.Name()), "Response") {
				return f.Message
			}
		}
	}
	
	return nil
}

// toPascalCase converts a string to PascalCase
func toPascalCase(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	s = strings.Title(s)
	return strings.ReplaceAll(s, " ", "")
}

// toCamelCase converts a string to camelCase
func toCamelCase(s string) string {
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		parts[i] = strings.Title(parts[i])
	}
	return strings.Join(parts, "")
}

// toSnakeCase converts a string to snake_case
func toSnakeCase(s string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")

	snake := matchFirstCap.ReplaceAllString(s, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}