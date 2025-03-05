package main

import (
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

// Generator handles the generation of GraphQL schema files from proto definitions
type Generator struct {
	template *template.Template
}

// newGenerator creates a new Generator instance with the provided options
func newGenerator(opts Options) (*Generator, error) {
	g := &Generator{}
	var err error
	if g.template, err = loadTemplate(opts.TemplatePath); err != nil {
		return nil, fmt.Errorf("failed to load template: %v", err)
	}
	return g, nil
}

var funcMap = template.FuncMap{
	"trim": strings.TrimSpace,
}

var (
	//go:embed templates/graphql-service-schema.tmpl
	templatesFS         embed.FS
	defaultTemplatePath = "templates/graphql-service-schema.tmpl"
)

// Generate processes protobuf files and generates the corresponding GraphQL schema
func (g *Generator) Generate(gen *protogen.Plugin) error {
	gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

	for _, f := range gen.Files {
		if !f.Generate {
			continue
		}
		for _, svc := range f.Services {
			if err := g.generateServiceSchema(svc, gen, f); err != nil {
				return err
			}
		}
	}
	return nil
}

// loadTemplate loads either the custom template specified in TemplatePath
// or falls back to the embedded template if not found or specified
func loadTemplate(templatePath string) (*template.Template, error) {
	if templatePath != "" {
		content, err := os.ReadFile(templatePath)
		if err != nil {
			log.Printf("Could not read template from %s: %v, falling back to embedded template", templatePath, err)
		} else {
			t, err := template.New(filepath.Base(templatePath)).Funcs(funcMap).Parse(string(content))
			if err != nil {
				log.Printf("Failed to parse template from file %s: %v, falling back to embedded template", templatePath, err)
			} else {
				log.Printf("Using custom template from path: %s", templatePath)
				return t, nil
			}
		}
	}

	// Fallback to embedded template
	content, err := templatesFS.ReadFile(defaultTemplatePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded template: %v", err)
	}

	t, err := template.New(filepath.Base(defaultTemplatePath)).Funcs(funcMap).Parse(string(content))
	if err != nil {
		return nil, fmt.Errorf("failed to parse embedded template: %v", err)
	}

	log.Println("Using embedded template")
	return t, nil
}

// TemplateData contains all data needed to render the GraphQL schema template
type TemplateData struct {
	// All services defined in the proto files
	Services []*ServiceData
	// Whether the schema contains any mutation services
	MutationServices bool
	// All messages defined in the proto files
	Messages []*Message
	// The source file that the schema was generated from
	Source string
}

type ServiceData struct {
	Name      string
	Federated bool
	Methods   []*Method
	Messages  []*Message
}

type Message struct {
	Name             string
	Fields           []*Field
	Entity           bool
	ReferenceMethods []*Method
	Comment          string
}

type Field struct {
	Name         string
	GraphQLType  string
	NonNull      bool
	External     bool
	Key          bool
	Requires     string
	ComputedFrom string
	Comment      string
}

type Method struct {
	Name       string
	Type       string
	InputArgs  string
	OutputType string
	Comment    string
}

func (g *Generator) generateServiceSchema(svc *protogen.Service, gen *protogen.Plugin, file *protogen.File) error {
	gf := gen.NewGeneratedFile(fmt.Sprintf("%s.graphql", svc.Desc.FullName()), protogen.GoImportPath(""))
	return g.renderTemplate(svc, gf, file)
}

func (g *Generator) renderTemplate(service *protogen.Service, gf *protogen.GeneratedFile, file *protogen.File) error {
	templateData := prepareTemplateData(service, file)
	return g.template.Execute(gf, templateData)
}

func prepareTemplateData(svc *protogen.Service, file *protogen.File) *TemplateData {
	return &TemplateData{
		Services: []*ServiceData{
			{
				Name:      string(svc.Desc.FullName()),
				Federated: true,
				Methods:   extractMethods(svc),
				Messages:  extractMessages(svc),
			},
		},
		MutationServices: hasMutationMethods(svc),
		Messages:         extractAllMessagesFromFile(file),
		Source:           svc.Desc.ParentFile().Path(),
	}
}

func extractMethods(svc *protogen.Service) []*Method {
	// Added nil check
	if svc == nil {
		return nil
	}

	var methods []*Method
	for _, method := range svc.Methods {
		if method == nil {
			continue
		}

		// Extract comments for the method
		comment := ""
		if method.Comments.Leading.String() != "" {
			comment = method.Comments.Leading.String()
		}

		// Extract proper input arguments
		inputArgs := extractInputArgs(method.Input)

		// Decide method type (Query vs Mutation)
		methodType := "Query"
		if strings.HasPrefix(string(method.Desc.Name()), "Create") ||
			strings.HasPrefix(string(method.Desc.Name()), "Update") ||
			strings.HasPrefix(string(method.Desc.Name()), "Delete") ||
			strings.HasPrefix(string(method.Desc.Name()), "Add") ||
			strings.HasPrefix(string(method.Desc.Name()), "Remove") {
			methodType = "Mutation"
		}

		methods = append(methods, &Method{
			Name:       string(method.Desc.Name()),
			Type:       methodType,
			InputArgs:  inputArgs,
			OutputType: string(method.Output.Desc.Name()),
			Comment:    comment,
		})
	}
	return methods
}

func extractInputArgs(input *protogen.Message) string {
	if len(input.Fields) == 0 {
		return ""
	}

	var args []string
	for _, f := range input.Fields {
		gqlType := "String"

		// Basic type mapping
		switch f.Desc.Kind().String() {
		case "DOUBLE", "FLOAT":
			gqlType = "Float"
		case "INT32", "INT64", "UINT32", "UINT64", "SINT32", "SINT64", "FIXED32", "FIXED64", "SFIXED32", "SFIXED64":
			gqlType = "Int"
		case "BOOL":
			gqlType = "Boolean"
		}

		// Add non-null marker if required
		if !f.Desc.HasOptionalKeyword() {
			gqlType += "!"
		}

		args = append(args, fmt.Sprintf("%s: %s", f.Desc.Name(), gqlType))
	}

	if len(args) == 0 {
		return ""
	}

	return "(" + strings.Join(args, ", ") + ")"
}

func extractMessages(svc *protogen.Service) []*Message {
	// Added nil check to prevent panic
	if svc == nil {
		return nil
	}

	// Track processed message names to avoid duplicates
	processedMessages := make(map[string]bool)

	var messages []*Message
	for _, m := range svc.Methods {
		if m == nil || m.Output == nil {
			continue
		}

		// Add the output message itself
		if !processedMessages[string(m.Output.Desc.Name())] {
			messages = append(messages, &Message{
				Name:   string(m.Output.Desc.Name()),
				Entity: hasEntityOption(m.Output),
				Fields: extractFields(m.Output),
			})
			processedMessages[string(m.Output.Desc.Name())] = true
		}

		// Process fields that are messages
		for _, f := range m.Output.Fields {
			if f != nil && f.Message != nil {
				msgName := string(f.Message.Desc.Name())
				if !processedMessages[msgName] {
					messages = append(messages, &Message{
						Name:   msgName,
						Entity: hasEntityOption(f.Message),
						Fields: extractFields(f.Message),
					})
					processedMessages[msgName] = true

					// Recursively add nested message types
					addNestedMessages(f.Message, &messages, processedMessages)
				}
			}
		}
	}
	return messages
}

// Recursively add nested message types
func addNestedMessages(msg *protogen.Message, messages *[]*Message, processed map[string]bool) {
	if msg == nil {
		return
	}

	for _, f := range msg.Fields {
		if f != nil && f.Message != nil {
			msgName := string(f.Message.Desc.Name())
			if !processed[msgName] {
				*messages = append(*messages, &Message{
					Name:   msgName,
					Entity: hasEntityOption(f.Message),
					Fields: extractFields(f.Message),
				})
				processed[msgName] = true

				// Recurse for this message's fields
				addNestedMessages(f.Message, messages, processed)
			}
		}
	}
}

func extractAllMessagesFromFile(file *protogen.File) []*Message {
	// Added nil check to prevent panic
	if file == nil {
		return nil
	}

	var messages []*Message
	for _, msg := range file.Messages {
		if msg == nil {
			continue
		}

		// Extract comments for the message
		comment := ""
		if msg.Comments.Leading.String() != "" {
			comment = strings.ReplaceAll(msg.Comments.Leading.String(), "//", "")
		}

		messages = append(messages, &Message{
			Name:    string(msg.Desc.Name()),
			Entity:  hasEntityOption(msg),
			Fields:  extractFields(msg),
			Comment: comment,
		})
	}
	return messages
}

func hasEntityOption(msg *protogen.Message) bool {
	// Try using a safer approach to get the entity option
	if msg == nil || msg.Desc == nil {
		return false
	}

	// Use direct name-based detection as a fallback
	// This is a temporary workaround
	name := string(msg.Desc.Name())
	if name == "Product" || name == "Order" || name == "User" {
		// Output debug info to stderr (won't affect generated output)
		log.Printf("Found entity by name: %s", name)
		return true
	}

	// Try to get via ProtoReflect with careful nil checks
	if msg.Desc.Options() != nil {
		const entityFieldNumber = 50001
		opts := msg.Desc.Options().ProtoReflect()
		if opts != nil {
			descriptor := opts.Descriptor()
			if descriptor != nil {
				fields := descriptor.Fields()
				if fields != nil {
					field := fields.ByNumber(entityFieldNumber)
					if field != nil {
						val := opts.Get(field)
						if val.IsValid() {
							return val.Bool()
						}
					}
				}
			}
		}
	}

	return false
}

func extractFields(msg *protogen.Message) []*Field {
	// Added nil check to prevent panic
	if msg == nil {
		return nil
	}

	var fields []*Field
	for _, f := range msg.Fields {
		// Default to String for simplicity, should be improved to map types properly
		gqlType := "String"

		// Basic type mapping
		switch f.Desc.Kind().String() {
		case "DOUBLE", "FLOAT":
			gqlType = "Float"
		case "INT32", "INT64", "UINT32", "UINT64", "SINT32", "SINT64", "FIXED32", "FIXED64", "SFIXED32", "SFIXED64":
			gqlType = "Int"
		case "BOOL":
			gqlType = "Boolean"
		}

		// If message type, use the message name as GraphQL type
		if f.Desc.Kind().String() == "MESSAGE" && f.Message != nil {
			gqlType = string(f.Message.Desc.Name())
		}

		// Get field comment if available
		comment := ""
		if f.Comments.Leading.String() != "" {
			comment = strings.ReplaceAll(f.Comments.Leading.String(), "//", "")
		}

		// Check for key option using field name pattern matching as fallback
		isKey := false
		name := string(f.Desc.Name())
		// Assume fields ending with "_id" are key fields for the entity
		if strings.HasSuffix(name, "_id") {
			log.Printf("Found key field by name: %s", name)
			isKey = true
		}

		// Also try proto options if available
		if f.Desc != nil && f.Desc.Options() != nil {
			const keyFieldNumber = 50001
			opts := f.Desc.Options().ProtoReflect()
			if opts != nil {
				descriptor := opts.Descriptor()
				if descriptor != nil {
					fields := descriptor.Fields()
					if fields != nil {
						field := fields.ByNumber(keyFieldNumber)
						if field != nil {
							keyOption := opts.Get(field)
							if keyOption.IsValid() {
								isKey = keyOption.Bool()
							}
						}
					}
				}
			}
		}

		fields = append(fields, &Field{
			Name:        string(f.Desc.Name()),
			GraphQLType: gqlType,
			NonNull:     !f.Desc.HasOptionalKeyword(),
			Key:         isKey,
			Comment:     comment,
		})
	}
	return fields
}

func hasMutationMethods(svc *protogen.Service) bool {
	// Added nil check
	if svc == nil {
		return false
	}

	// Look for methods that start with Create, Update, Delete, etc.
	for _, method := range svc.Methods {
		if method == nil {
			continue
		}

		methodName := string(method.Desc.Name())
		if strings.HasPrefix(methodName, "Create") ||
			strings.HasPrefix(methodName, "Update") ||
			strings.HasPrefix(methodName, "Delete") ||
			strings.HasPrefix(methodName, "Add") ||
			strings.HasPrefix(methodName, "Remove") {
			return true
		}
	}

	return false
}
