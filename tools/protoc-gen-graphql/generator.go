package main

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

// Generator handles the generation of GraphQL schema files from proto definitions
type Generator struct {
	TemplatePath string
}

func newGenerator(templatePath string) *Generator {
	return &Generator{TemplatePath: templatePath}
}

// Generate processes protobuf files and generates the corresponding GraphQL schema
func (g *Generator) Generate(gen *protogen.Plugin) error {
	gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

	for _, f := range gen.Files {
		if !f.Generate {
			continue
		}
		for _, svc := range f.Services {
			if err := g.generateServiceSchema(svc, gen); err != nil {
				return err
			}
		}
	}
	return nil
}

// TemplateData contains all data needed to render the GraphQL schema template
type TemplateData struct {
	Services         []*ServiceData
	MutationServices bool
	Messages         []*Message
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
}

type Field struct {
	Name         string
	GraphQLType  string
	NonNull      bool
	External     bool
	Key          bool
	Requires     string
	ComputedFrom string
}

type Method struct {
	Name       string
	Type       string
	InputArgs  string
	OutputType string
}

func (g *Generator) generateServiceSchema(svc *protogen.Service, gen *protogen.Plugin) error {
	gf := gen.NewGeneratedFile(fmt.Sprintf("%s.graphql", svc.Desc.FullName()), protogen.GoImportPath(""))
	return g.renderTemplate(svc, gf)
}

func (g *Generator) renderTemplate(service *protogen.Service, gf *protogen.GeneratedFile) error {
	content, err := os.ReadFile(g.TemplatePath)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %v", g.TemplatePath, err)
	}
	t, err := template.New(filepath.Base(g.TemplatePath)).Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}
	templateData := prepareTemplateData(service)
	return t.Execute(gf, templateData)
}

func prepareTemplateData(svc *protogen.Service) *TemplateData {
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
		Messages:         extractMessages(svc),
	}
}

func extractMethods(svc *protogen.Service) []*Method {
	var methods []*Method
	for _, method := range svc.Methods {
		methods = append(methods, &Method{
			Name:       string(method.Desc.Name()),
			Type:       "Query",
			InputArgs:  "(id: ID!)",
			OutputType: string(method.Output.Desc.Name()),
		})
	}
	return methods
}

func extractMessages(svc *protogen.Service) []*Message {
	var messages []*Message
	for _, m := range svc.Methods {
		for _, f := range m.Output.Fields {
			if f.Message != nil {
				messages = append(messages, &Message{
					Name:   string(f.Message.Desc.Name()),
					Entity: true,
					Fields: extractFields(f.Message),
				})
			}
		}
	}
	return messages
}

func extractFields(msg *protogen.Message) []*Field {
	var fields []*Field
	for _, f := range msg.Fields {
		fields = append(fields, &Field{
			Name:        string(f.Desc.Name()),
			GraphQLType: "String",
			NonNull:     true,
		})
	}
	return fields
}

func hasMutationMethods(svc *protogen.Service) bool {
	return false
}
