package gqltypes

// Schema represents a GraphQL schema
type Schema struct {
	ServiceName  string
	RootQuery    *Type
	RootMutation *Type
	Types        []*Type
	Inputs       []*Input
	Enums        []*Enum
}

// Type represents a GraphQL type
type Type struct {
	Name              string
	Fields            []*Field
	Comment           string
	IsFederatedEntity bool   // For @key directive
	KeyFields         string // Fields used in @key
}

// Field represents a field in a GraphQL type
type Field struct {
	Name       string
	Type       string
	Comment    string
	Inputs     []*Input
	IsRequired bool
}

// Input represents a GraphQL input type
type Input struct {
	Name    string
	Fields  []*Field
	Comment string
}

// Enum represents a GraphQL enum type
type Enum struct {
	Name    string
	Options []*EnumOption
	Comment string
}

// EnumOption represents an option in a GraphQL enum
type EnumOption struct {
	Name    string
	Comment string
}
