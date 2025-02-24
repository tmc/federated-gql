package main

import (
	"embed"
	"fmt"
	"regexp"
	"strings"
	"text/template"

	"github.com/fraser-isbester/federated-graphql/tools/protoc-gen-graphql/gqltypes"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

type Generator struct {
	TemplateDir string
	services    map[string]*protogen.Service
	schemas     map[string]*gqltypes.Schema
}

func newGenerator(templateDir string) *Generator {
	return &Generator{
		TemplateDir: templateDir,
		services:    make(map[string]*protogen.Service),
		schemas:     make(map[string]*gqltypes.Schema),
	}
}

func (g *Generator) Generate(gen *protogen.Plugin) error {
	gen.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
	if err := g.walkSchemas(gen); err != nil {
		return err
	}
	return g.printSchemas(gen)
}

//go:embed templates/*
var defaultTemplates embed.FS

func (g *Generator) renderTemplate(templateName string, service *protogen.Service, gf *protogen.GeneratedFile, funcMap template.FuncMap) error {
	// Read template content directly from embedded FS
	content, err := defaultTemplates.ReadFile("templates/" + templateName)
	if err != nil {
		return fmt.Errorf("failed to read template: %v", err)
	}

	// Create and parse template
	t := template.New(templateName).Funcs(funcMap)
	t, err = t.Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse template: %v", err)
	}

	// Execute template
	return t.Execute(gf, service)
}

func (g *Generator) walkSchemas(gen *protogen.Plugin) error {
	for _, f := range gen.Files {
		if !f.Generate {
			continue
		}
		for _, svc := range f.Services {
			for _, m := range svc.Methods {
				g.walkMethod(m)
			}
		}
	}
	return nil
}

func (g *Generator) walkMethod(m *protogen.Method) {

	svcFullName := string(m.Parent.Desc.FullName())
	if _, ok := g.services[svcFullName]; !ok {
		g.services[svcFullName] = m.Parent
		schema := newSchema()
		// Set service name from the package name
		pkgParts := strings.Split(svcFullName, ".")
		schema.ServiceName = strings.TrimSuffix(pkgParts[0], "v1")
		g.schemas[svcFullName] = schema
	}

	if _, ok := g.services[string(m.Parent.Desc.FullName())]; !ok {
		g.services[string(m.Parent.Desc.FullName())] = m.Parent
		g.schemas[string(m.Parent.Desc.FullName())] = newSchema()
	}
	schema := g.schemas[string(m.Parent.Desc.FullName())]

	// Add input type
	input := g.getInputType(schema, m.Input)

	// Add output type
	output := g.getOutputType(schema, m.Output)

	// Add method to Query if it's a getter
	isQuery := strings.HasPrefix(strings.ToLower(string(m.Desc.Name())), "get") ||
		strings.HasPrefix(strings.ToLower(string(m.Desc.Name())), "list") ||
		strings.HasPrefix(strings.ToLower(string(m.Desc.Name())), "search")

	field := &gqltypes.Field{
		Name:    g.getMethodName(m),
		Comment: cleanComment(string(m.Comments.Leading)),
		Inputs:  []*gqltypes.Input{input},
		Type:    output.Name,
	}

	if isQuery {
		schema.RootQuery.Fields = append(schema.RootQuery.Fields, field)
	} else {
		schema.RootMutation.Fields = append(schema.RootMutation.Fields, field)
	}
}

func newSchema() *gqltypes.Schema {
	return &gqltypes.Schema{
		RootQuery:    &gqltypes.Type{Name: "Query"},
		RootMutation: &gqltypes.Type{Name: "Mutation"},
		Types:        make([]*gqltypes.Type, 0),
		Inputs:       make([]*gqltypes.Input, 0),
		Enums:        make([]*gqltypes.Enum, 0),
	}
}

func (g *Generator) getMethodName(m *protogen.Method) string {
	n := string(m.Desc.Name())
	return strings.ToLower(string(n[0])) + n[1:]
}

func (g *Generator) getInputType(s *gqltypes.Schema, m *protogen.Message) *gqltypes.Input {
	name := g.getInputTypeName(m)

	// Check if input type already exists
	for _, input := range s.Inputs {
		if input.Name == name {
			return input
		}
	}

	input := &gqltypes.Input{
		Name:    name,
		Fields:  make([]*gqltypes.Field, 0),
		Comment: cleanComment(string(m.Comments.Leading)),
	}

	for _, f := range m.Fields {
		field := &gqltypes.Field{
			Name:       string(f.Desc.JSONName()),
			Type:       g.getFieldType(s, f),
			Comment:    cleanComment(string(f.Comments.Leading)),
			IsRequired: f.Desc.HasPresence(),
		}
		input.Fields = append(input.Fields, field)
	}

	s.Inputs = append(s.Inputs, input)
	return input
}

func (g *Generator) getOutputType(s *gqltypes.Schema, m *protogen.Message) *gqltypes.Type {
	name := string(m.Desc.Name())

	// Check if type already exists
	for _, t := range s.Types {
		if t.Name == name {
			return t
		}
	}

	output := &gqltypes.Type{
		Name:    name,
		Fields:  make([]*gqltypes.Field, 0),
		Comment: cleanComment(string(m.Comments.Leading)),
	}

	for _, f := range m.Fields {
		field := &gqltypes.Field{
			Name:       string(f.Desc.JSONName()),
			Type:       g.getFieldType(s, f),
			Comment:    cleanComment(string(f.Comments.Leading)),
			IsRequired: f.Desc.HasPresence(),
		}
		output.Fields = append(output.Fields, field)
	}

	// Check if this type should be federated
	output.ShouldFederate()

	s.Types = append(s.Types, output)
	return output
}
func (g *Generator) getFieldType(s *gqltypes.Schema, f *protogen.Field) string {
	if f.Message != nil {
		return g.getOutputType(s, f.Message).Name
	}
	if f.Enum != nil {
		return g.getEnumType(s, f.Enum)
	}
	return gqltypes.ProtoKindToGraphQL(f.Desc.Kind(), f.Desc.IsList(), f.Desc.HasPresence())
}

func (g *Generator) getEnumType(s *gqltypes.Schema, e *protogen.Enum) string {
	name := string(e.Desc.Name())

	// Check if enum already exists
	for _, enum := range s.Enums {
		if enum.Name == name {
			return name
		}
	}

	enum := gqltypes.ProtoEnumToGraphQL(e.Desc)
	s.Enums = append(s.Enums, enum)
	return name
}

func (g *Generator) getInputTypeName(m *protogen.Message) string {
	n := string(m.Desc.Name())
	n = strings.TrimSuffix(n, "Request")
	return n + "Input"
}

func (g *Generator) printSchemas(plugin *protogen.Plugin) error {
	for _, f := range plugin.Files {
		if !f.Generate {
			continue
		}
		for _, svc := range f.Services {
			if err := g.printServiceSchema(svc, plugin); err != nil {
				return err
			}
		}
	}
	return nil
}

var commentRe = regexp.MustCompile(`\n\s*//\s*`)

func cleanComment(comment string) string {
	if comment == "" {
		return ""
	}
	// Remove leading comment markers and whitespace
	comment = strings.TrimLeft(comment, "*/\n ")
	// Replace newline comments with spaces
	comment = commentRe.ReplaceAllString(comment, " ")
	// Clean up any remaining whitespace
	comment = strings.TrimSpace(comment)
	return comment
}

func (g *Generator) printServiceSchema(svc *protogen.Service, gen *protogen.Plugin) error {
	serviceName := string(svc.Desc.FullName())
	gf := gen.NewGeneratedFile(fmt.Sprintf("%s.graphql", serviceName), protogen.GoImportPath(serviceName))

	return g.renderTemplate("graphql-service-schema.tmpl", svc, gf, template.FuncMap{
		"schema": func(svc *protogen.Service) *gqltypes.Schema {
			return g.schemas[string(svc.Desc.FullName())]
		},
		"lower":      strings.ToLower,
		"trimPrefix": strings.TrimPrefix,
		"trimSuffix": strings.TrimSuffix,
		"hasField": func(t *gqltypes.Type, fieldPrefix string) bool {
			for _, f := range t.Fields {
				if strings.HasSuffix(strings.ToLower(f.Name), strings.ToLower(fieldPrefix)) {
					return true
				}
			}
			return false
		},
		"idField": func(t *gqltypes.Type) string {
			for _, f := range t.Fields {
				if strings.HasSuffix(strings.ToLower(f.Name), "id") {
					return f.Name
				}
			}
			return ""
		},
		"isIDField": func(name string) bool {
			return strings.HasSuffix(strings.ToLower(name), "id")
		},
	})
}
